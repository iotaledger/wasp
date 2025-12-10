// Package consensus implements consensus functionality for IOTA Smart Contracts.
// A single instance of it.
//
// We move all the synchronization logic to separate objects (upon_...). They are
// responsible for waiting specific data and then triggering the next state action
// once. This way we hope to solve a lot of race conditions gracefully. The `upon`
// predicates and the corresponding done functions should not depend on each other.
// If some data is needed at several places, it should be passed to several predicates.
package consensus

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"fortio.org/safecast"
	"github.com/minio/blake2b-simd"
	"github.com/samber/lo"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/chain/consensus/batchproposal"
	"github.com/iotaledger/wasp/v2/packages/chain/distsign"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/aba/mostefaoui"
	"github.com/iotaledger/wasp/v2/packages/gpa/acs"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	"github.com/iotaledger/wasp/v2/packages/gpa/asyncdistkeygen/nonce"
	"github.com/iotaledger/wasp/v2/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/v2/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
	"github.com/iotaledger/wasp/v2/packages/vm/vmtxbuilder"
)

type OutputStatus byte

func (os OutputStatus) String() string {
	switch os {
	case Running:
		return "Running"
	case Completed:
		return "Completed"
	case Skipped:
		return "Skipped"
	default:
		return fmt.Sprintf("Unexpected-%v", byte(os))
	}
}

const (
	Running   OutputStatus = iota // Instance is still running.
	Completed                     // Consensus reached, TX is prepared for publication.
	Skipped                       // Consensus reached, no TX should be posted for this LogIndex.
)

type Output struct {
	Status     OutputStatus
	Terminated bool
	//
	// Requests for other components.
	NeedMempoolProposal       *isc.StateAnchor  // Requests for the mempool are needed for this Base Alias Output.
	NeedMempoolRequests       []*isc.RequestRef // Request payloads are needed from mempool for this IDs/Hash.
	NeedStateMgrStateProposal *isc.StateAnchor  // Query for a proposal for Virtual State (it will go to the batch proposal).
	NeedStateMgrDecidedState  *isc.StateAnchor  // Query for a decided Virtual State to be used by VM.
	NeedStateMgrSaveBlock     state.StateDraft  // Ask StateMgr to save the produced block.
	NeedNodeConnL1Info        *isc.StateAnchor  // Ask NodeConn for the L1Info related to this anchor.
	NeedVMResult              *vm.VMTask        // VM Result is needed for this (agreed) batch.
	//
	// Following is the final result.
	// All the fields are filled, if State == Completed.
	Result *Result
}

type Result struct {
	DecidedAnchor *isc.StateAnchor              // The consumed state anchor.
	Transaction   *iotasigner.SignedTransaction // The TX for committing the block.
	Block         state.Block                   // The state diff produced.
}

func (r *Result) String() string {
	return fmt.Sprintf(
		"{cons.Result, txDigest=%s, baseAnchor=%v, outBlockHash=%v}",
		lo.Must(r.Transaction.Digest()),
		r.DecidedAnchor,
		r.Block.Hash(),
	)
}

type Consensus struct {
	chainID                 isc.ChainID
	chainStore              state.Store
	edSuite                 suites.Suite    // For signatures.
	blsSuite                suites.Suite    // For randomness only.
	dkShare                 tcrypto.DKShare // The current committee's keys.
	rotateTo                *iotago.Address // If non-nil and differs from the dkShare, then rotation is suggested.
	processorCache          *processors.Config
	nodeIDs                 []gpa.NodeID
	me                      gpa.NodeID
	f                       int
	asGPA                   gpa.GPA
	distributedSignature    *distsign.DistributedSignature
	acs                     *acs.ACS
	subMempool              *SyncMempool              // Mempool.
	subStateMgr             *SyncStateMgr             // StateMgr.
	subNodeconn             *SyncNodeconn             // Synchronization with the NodeConn.
	subDistributedSignature *SyncDistributedSignature // Distributed Schnorr Signature.
	subACS                  *SyncACS                  // Asynchronous Common Subset.
	subRND                  *SyncRND                  // Randomness.
	subVM                   *SyncVM                   // Virtual Machine.
	subTX                   *SyncTX                   // Building final TX.
	term                    *termCondition            // To detect, when this instance can be terminated.
	output                  *Output
	validatorAgentID        isc.AgentID
	log                     log.Logger
}

var _ gpa.GPA = &Consensus{}

func New(
	chainID isc.ChainID,
	chainStore state.Store,
	me gpa.NodeID,
	mySK *cryptolib.PrivateKey,
	dkShare tcrypto.DKShare,
	rotateTo *iotago.Address,
	processorCache *processors.Config,
	instID []byte,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	validatorAgentID isc.AgentID,
	log log.Logger,
) *Consensus {
	edSuite := tcrypto.DefaultEd25519Suite()
	blsSuite := tcrypto.DefaultBLSSuite()

	dkShareNodePubKeys := dkShare.GetNodePubKeys()
	nodeIDs := make([]gpa.NodeID, len(dkShareNodePubKeys))
	nodePKs := map[gpa.NodeID]kyber.Point{}
	for i := range dkShareNodePubKeys {
		var err error
		nodeIDs[i] = nodeIDFromPubKey(dkShareNodePubKeys[i])
		nodePKs[nodeIDs[i]], err = dkShareNodePubKeys[i].AsKyberPoint()
		if err != nil {
			panic(fmt.Errorf("cannot convert nodePK[%v] to kyber.Point: %w", i, err))
		}
	}

	f := len(dkShareNodePubKeys) - int(dkShare.GetT())
	myKyberKeys, err := mySK.AsKyberKeyPair()
	if err != nil {
		panic(fmt.Errorf("cannot convert node's SK to kyber.Scalar: %w", err))
	}
	longTermDKS := dkShare.DSS()
	acsLog := log.NewChildLogger("ACS")
	acsCCInstFunc := func(nodeID gpa.NodeID, round int) *semi.CCSemi {
		var roundBin [4]byte
		roundU32, err := safecast.Convert[uint32](round)
		if err != nil {
			panic("round overflows uint32")
		}
		binary.BigEndian.PutUint32(roundBin[:], roundU32)
		sid := hashing.HashDataBlake2b(instID, nodeID[:], roundBin[:]).Bytes()
		realCC := blssig.New(blsSuite, nodeIDs, dkShare.BLSCommits(), dkShare.BLSPriShare(), int(dkShare.BLSThreshold()), me, sid, acsLog)
		return semi.New(round, realCC)
	}
	c := &Consensus{
		chainID:              chainID,
		chainStore:           chainStore,
		edSuite:              edSuite,
		blsSuite:             blsSuite,
		dkShare:              dkShare,
		rotateTo:             rotateTo,
		processorCache:       processorCache,
		nodeIDs:              nodeIDs,
		me:                   me,
		f:                    f,
		distributedSignature: distsign.New(edSuite, nodeIDs, nodePKs, f, me, myKyberKeys.Private, longTermDKS, log.NewChildLogger("DistributedSignature")),
		acs:                  acs.New(nodeIDs, me, f, acsCCInstFunc, acsLog),
		output:               &Output{Status: Running},
		log:                  log,
		validatorAgentID:     validatorAgentID,
	}
	c.asGPA = gpa.NewOwnHandler(me, c)
	c.subMempool = NewSyncMempool(c)
	c.subStateMgr = NewSyncStateMgr(c)
	c.subNodeconn = NewSyncNodeconn(c)
	c.subDistributedSignature = NewSyncDistributedSignature(c)
	c.subACS = NewSyncACS(c)
	c.subRND = NewSyncRND(int(dkShare.BLSThreshold()), c)
	c.subVM = NewSyncVM(c)
	c.subTX = NewSyncTX(c)
	c.term = newTermCondition(
		c.uponTerminationCondition,
	)
	return c
}

func (c *Consensus) AsGPA() gpa.GPA {
	return c.asGPA
}

func (c *Consensus) Input(input gpa.Input) []gpa.MessageOut {
	switch input := input.(type) {
	case *inputTimeData:
		// ignore this to filter out ridiculously excessive logging
	default:
		c.log.LogDebugf("Input %T: %+v", input, input)
	}

	switch input := input.(type) {
	case *inputProposal:
		c.log.LogInfof("Consensus started, received %v", input.String())
		return slices.Concat(
			c.subNodeconn.HaveInputAnchor(input.baseAnchor),
			c.subMempool.BaseAnchorReceived(input.baseAnchor),
			c.subStateMgr.ProposedBaseAnchorReceived(input.baseAnchor),
			c.subDistributedSignature.InitialInputReceived(),
		)
	case *inputRotateTo:
		// We can update the rotation address while consensus is running.
		// New value will be used, if decision has not been made yet.
		c.rotateTo = input.address
		return nil
	case *inputMempoolProposal:
		return c.subMempool.ProposalReceived(input.requestRefs)
	case *inputMempoolRequests:
		return c.subMempool.RequestsReceived(input.requests)
	case *inputStateMgrProposalConfirmed:
		return c.subStateMgr.StateProposalConfirmedByStateMgr()
	case *inputStateMgrDecidedVirtualState:
		return c.subStateMgr.DecidedVirtualStateReceived(input.chainState)
	case *inputStateMgrBlockSaved:
		return c.subStateMgr.BlockSaved(input.block)
	case *inputTimeData:
		return c.subACS.TimeDataReceived(input.timeData)
	case *inputL1Info:
		return c.subNodeconn.HaveL1Info(input.gasCoins, input.l1params)
	case *inputVMResult:
		return c.subVM.VMResultReceived(input.task)
	}
	panic(fmt.Errorf("unexpected input: %v", input))
}

// Message implements the gpa.GPA interface.
// Here we route all the messages.
func (c *Consensus) Message(msg gpa.MessageIn[any]) []gpa.MessageOut {
	switch msgT := msg.Payload.(type) {
	case msgBLSPartialSig:
		return c.subRND.BLSPartialSigReceived(msg.Sender, msgT.partialSig)
	case gpa.PayloadWithKey[int, mostefaoui.MsgDone]:
		msgs := c.acs.HandleABAMsgDone(msgT.Key, gpa.NewMessageIn(msg.Sender, msgT.Payload))
		return slices.Concat(msgs, c.subACS.ACSOutputReceived(c.acs.Output()))
	case gpa.PayloadWithKey[int, mostefaoui.MsgVote]:
		msgs := c.acs.HandleABAMsgVote(msgT.Key, gpa.NewMessageIn(msg.Sender, msgT.Payload))
		return slices.Concat(msgs, c.subACS.ACSOutputReceived(c.acs.Output()))
	case gpa.PayloadWithKey[int, gpa.PayloadWithKey[int, blssig.MsgSigShare]]:
		abaIndex := msgT.Key
		ccIndex := msgT.Payload.Key
		return c.acs.HandleCCMsgSigShare(abaIndex, ccIndex, gpa.NewMessageIn(msg.Sender, msgT.Payload.Payload))
	case gpa.PayloadWithKey[int, bracha.MsgBracha]:
		switch msgT.SubsystemID {
		case acs.SubsystemID:
			msgs := c.acs.HandleRBCMsgBracha(msgT.Key, gpa.NewMessageIn(msg.Sender, msgT.Payload))
			return slices.Concat(msgs, c.subACS.ACSOutputReceived(c.acs.Output()))
		case nonce.SubsystemID:
			msgs := c.distributedSignature.HandleACSSMsgBracha(msgT.Key, gpa.NewMessageIn(msg.Sender, msgT.Payload))
			return slices.Concat(msgs, c.subDistributedSignature.DistributedSignatureReady(c.distributedSignature.Output()))
		default:
			c.log.LogErrorf("cannot select subsystem: unexpected subsystem ID: %s", msgT.SubsystemID)
		}
	case gpa.PayloadWithKey[int, acss.MsgVote]:
		msgs := c.distributedSignature.HandleACSSMsgVote(msgT.Key, gpa.NewMessageIn(msg.Sender, msgT.Payload))
		return slices.Concat(msgs, c.subDistributedSignature.DistributedSignatureReady(c.distributedSignature.Output()))
	case gpa.PayloadWithKey[int, acss.MsgImplicateRecover]:
		msgs := c.distributedSignature.HandleACSSMsgImplicateRecover(msgT.Key, gpa.NewMessageIn(msg.Sender, msgT.Payload))
		return slices.Concat(msgs, c.subDistributedSignature.DistributedSignatureReady(c.distributedSignature.Output()))
	case distsign.MsgPartialSig:
		msgs := c.distributedSignature.HandleMsgPartialSig(gpa.NewMessageIn(msg.Sender, msgT))
		return slices.Concat(msgs, c.subDistributedSignature.DistributedSignatureReady(c.distributedSignature.Output()))
	}
	panic(fmt.Errorf("unexpected message payload: %#v", msg.Payload))
}

func (c *Consensus) Output() gpa.Output {
	return c.output // Always non-nil.
}

func (c *Consensus) StatusString() string {
	// We con't include RND here, maybe that's less important, and visible from the VM status.
	return fmt.Sprintf("{consImpl⟨%v⟩,%v,%v,%v,%v,%v,%v,%v}",
		c.output.Status,
		c.subStateMgr.String(),
		c.subMempool.String(),
		c.subNodeconn.String(),
		c.subDistributedSignature.String(),
		c.subACS.String(),
		c.subVM.String(),
		c.subTX.String(),
	)
}

////////////////////////////////////////////////////////////////////////////////
// MP -- MemPool

func (c *Consensus) uponMempoolProposalInputsReady(baseAnchor *isc.StateAnchor) []gpa.MessageOut {
	if baseAnchor == nil {
		// If the base Anchor is nil, we are not going to propose any requests.
		return c.subMempool.ProposalReceived([]*isc.RequestRef{})
	}
	c.output.NeedMempoolProposal = baseAnchor
	return nil
}

func (c *Consensus) uponMempoolProposalReceived(requestRefs []*isc.RequestRef) []gpa.MessageOut {
	c.output.NeedMempoolProposal = nil
	return slices.Concat(
		c.subACS.MempoolRequestsReceived(requestRefs),
		c.subNodeconn.HaveRequests(),
	)
}

func (c *Consensus) uponMempoolRequestsNeeded(requestRefs []*isc.RequestRef) []gpa.MessageOut {
	c.output.NeedMempoolRequests = requestRefs
	return nil
}

func (c *Consensus) uponMempoolRequestsReceived(requests []isc.Request) []gpa.MessageOut {
	c.output.NeedMempoolRequests = nil
	return c.subVM.RequestsReceived(requests)
}

////////////////////////////////////////////////////////////////////////////////
// SM -- StateManager

func (c *Consensus) uponStateMgrStateProposalQueryInputsReady(baseAnchor *isc.StateAnchor) []gpa.MessageOut {
	if baseAnchor == nil {
		// Don't wait for the state if no base Anchor is known.
		return c.subStateMgr.StateProposalConfirmedByStateMgr()
	}
	c.output.NeedStateMgrStateProposal = baseAnchor
	return nil
}

func (c *Consensus) uponStateMgrStateProposalReceived(proposedAnchor *isc.StateAnchor) []gpa.MessageOut {
	c.output.NeedStateMgrStateProposal = nil
	return slices.Concat(
		c.subACS.StateProposalReceived(proposedAnchor),
		c.subNodeconn.HaveState(),
	)
}

func (c *Consensus) uponStateMgrDecidedStateQueryInputsReady(decidedBaseAnchor *isc.StateAnchor) []gpa.MessageOut {
	c.output.NeedStateMgrDecidedState = decidedBaseAnchor
	return nil
}

func (c *Consensus) uponStateMgrDecidedStateReceived(chainState state.State) []gpa.MessageOut {
	c.output.NeedStateMgrDecidedState = nil
	return c.subVM.DecidedStateReceived(chainState)
}

func (c *Consensus) uponStateMgrSaveProducedBlockInputsReady(producedBlock state.StateDraft) []gpa.MessageOut {
	if producedBlock == nil {
		// Don't have a block to save in the case of self-governed rotation.
		// So mark it as saved immediately.
		return c.subStateMgr.BlockSaved(nil)
	}
	c.output.NeedStateMgrSaveBlock = producedBlock
	return nil
}

func (c *Consensus) uponStateMgrSaveProducedBlockDone(block state.Block) []gpa.MessageOut {
	c.output.NeedStateMgrSaveBlock = nil
	return c.subTX.BlockSaved(block)
}

////////////////////////////////////////////////////////////////////////////////
// NC

func (c *Consensus) uponNodeconnInputsReady(anchor *isc.StateAnchor) []gpa.MessageOut {
	if anchor == nil {
		c.log.LogDebugf("ACS got ⊥ as input, no L1 info can be fetched.")
		return c.subACS.L1InfoReceived([]*coin.CoinWithRef{}, nil)
	}
	c.output.NeedNodeConnL1Info = anchor
	return nil
}

func (c *Consensus) uponNodeconnOutputReady(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) []gpa.MessageOut {
	c.log.LogDebugf("L1 info received, gasCoins=%v, l1Params=%v", gasCoins, l1params)
	c.output.NeedNodeConnL1Info = nil
	return c.subACS.L1InfoReceived(gasCoins, l1params)
}

////////////////////////////////////////////////////////////////////////////////
// DistributedSignature

func (c *Consensus) uponDistributedSignatureInitialInputsReady() []gpa.MessageOut {
	c.log.LogDebugf("uponDistributedSignatureInitialInputsReady")
	subMsgs := c.distributedSignature.Input(distsign.NewInputStart())
	return slices.Concat(subMsgs, c.subDistributedSignature.DistributedSignatureReady(c.distributedSignature.Output()))
}

func (c *Consensus) uponDistributedSignatureIndexProposalReady(indexProposal []int) []gpa.MessageOut {
	c.log.LogDebugf("uponDistributedSignatureIndexProposalReady")
	return c.subACS.DistributedSignatureIndexProposalReceived(indexProposal)
}

func (c *Consensus) uponDistributedSignatureSigningInputsReceived(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) []gpa.MessageOut {
	c.log.LogDebugf("uponDistributedSignatureSigningInputsReceived(decidedIndexProposals=%+v, H(messageToSign)=%v)", decidedIndexProposals, hashing.HashDataBlake2b(messageToSign))
	distributedSignatureDecidedInput := distsign.NewInputDecided(decidedIndexProposals, messageToSign)
	subMsgs := c.distributedSignature.Input(distributedSignatureDecidedInput)
	return slices.Concat(subMsgs, c.subDistributedSignature.DistributedSignatureReady(c.distributedSignature.Output()))
}

func (c *Consensus) uponDistributedSignatureOutputReady(signature []byte) []gpa.MessageOut {
	c.log.LogDebugf("uponDistributedSignatureOutputReady")
	return c.subTX.SignatureReceived(signature)
}

////////////////////////////////////////////////////////////////////////////////
// ACS

func (c *Consensus) uponACSInputsReceived(
	baseAnchor *isc.StateAnchor, // Can be nil.
	requestRefs []*isc.RequestRef,
	distSignIndexProposal []int,
	timeData time.Time,
	gasCoins []*coin.CoinWithRef, // Can be nil.
	l1params *parameters.L1Params, // Can be nil.
) []gpa.MessageOut {
	rotateTo := c.rotateTo
	if rotateTo != nil && rotateTo.Equals(*c.dkShare.GetAddress().AsIotaAddress()) {
		// Do not propose to rotate to the existing committee.
		rotateTo = nil
	}
	batchProposal := batchproposal.NewBatchProposal(
		*c.dkShare.GetIndex(),
		baseAnchor, // Will be NIL in the case of ⊥ proposal.
		util.NewFixedSizeBitVector(c.dkShare.GetN()).SetBits(distSignIndexProposal),
		rotateTo,
		timeData,
		c.validatorAgentID,
		requestRefs, // Will be [] in the case of ⊥ proposal.
		gasCoins,    // Will be NIL in the case of ⊥ proposal.
		l1params,    // Will be NIL in the case of ⊥ proposal.
	)

	subMsgs := c.acs.Input(batchProposal.Bytes())
	return slices.Concat(subMsgs, c.subACS.ACSOutputReceived(c.acs.Output()))
}

func (c *Consensus) uponACSOutputReceived(outputValues map[gpa.NodeID][]byte) []gpa.MessageOut {
	aggr := batchproposal.AggregateBatchProposals(outputValues, c.nodeIDs, c.f, c.log)
	if aggr.ShouldBeSkipped() {
		// Cannot proceed with such proposals.
		// Have to retry the consensus after some time with the next log index.
		c.log.LogInfof("Terminating consensus with status=Skipped, there is no way to aggregate batch proposal.")
		c.output.Status = Skipped
		c.term.haveOutputProduced()
		return nil
	}
	bao := aggr.DecidedBaseAnchor()
	baoID := bao.GetObjectRef()
	reqs := aggr.DecidedRequestRefs()
	c.log.LogDebugf("ACS decision: baseAnchor=%v, requests=%v", bao, reqs)
	if aggr.DecidedRotateTo() != nil {
		c.log.LogDebugf("Will rotate to %v", aggr.DecidedRotateTo().ToHex())
		rotationPTB := vmtxbuilder.NewAnchorTransactionBuilder(bao.ISCPackage(), bao, c.dkShare.GetAddress())
		rotationPTB.RotationTransaction(aggr.DecidedRotateTo())
		rotationPTX := rotationPTB.BuildTransactionEssence(bao.GetStateMetadata(), 0)
		rotationTXD := c.makeTransactionData(&rotationPTX, aggr)
		rotationTXB := c.makeTransactionSigningBytes(rotationTXD)
		c.log.LogDebugf("Rotation TxDataBytes=%s", hex.EncodeToString(c.makeTransactionDataBytes(rotationTXD)))
		return slices.Concat(
			c.subTX.UnsignedTXReceived(rotationTXD),
			c.subTX.BlockSaved(nil),
			c.subTX.AnchorDecided(bao),
			c.subDistributedSignature.MessageToSignReceived(rotationTXB),
			c.subDistributedSignature.DecidedIndexProposalsReceived(aggr.DecidedDistributedSignatureIndexProposals()),
		)
	}
	return slices.Concat(
		c.subMempool.RequestsNeeded(reqs),
		c.subStateMgr.DecidedVirtualStateNeeded(bao),
		c.subVM.DecidedBatchProposalsReceived(aggr),
		c.subRND.CanProceed(baoID.Bytes()),
		c.subDistributedSignature.DecidedIndexProposalsReceived(aggr.DecidedDistributedSignatureIndexProposals()),
	)
}

func (c *Consensus) uponACSTerminated() {
	c.term.haveAcsTerminated()
}

////////////////////////////////////////////////////////////////////////////////
// RND

func (c *Consensus) uponRNDInputsReady(dataToSign []byte) []gpa.MessageOut {
	sigShare, err := c.dkShare.BLSSignShare(dataToSign)
	if err != nil {
		panic(fmt.Errorf("cannot sign share for randomness: %w", err))
	}
	return lo.Map(c.nodeIDs, func(nid gpa.NodeID, _ int) gpa.MessageOut {
		return newMsgBLSPartialSig(nid, sigShare)
	})
}

func (c *Consensus) uponRNDSigSharesReady(dataToSign []byte, partialSigs map[gpa.NodeID][]byte) (bool, []gpa.MessageOut) {
	partialSigArray := make([][]byte, 0, len(partialSigs))
	for nid := range partialSigs {
		partialSigArray = append(partialSigArray, partialSigs[nid])
	}
	sig, err := c.dkShare.BLSRecoverMasterSignature(partialSigArray, dataToSign)
	if err != nil {
		c.log.LogWarnf("Cannot reconstruct BLS signature from %v/%v sigShares: %v", len(partialSigs), c.dkShare.GetN(), err)
		return false, nil // Continue to wait for other sig shares.
	}
	return true, c.subVM.RandomnessReceived(hashing.HashDataBlake2b(sig.Signature.Bytes()))
}

////////////////////////////////////////////////////////////////////////////////
// VM

func (c *Consensus) uponVMInputsReceived(aggregatedProposals *batchproposal.AggregatedBatchProposals, randomness *hashing.HashValue, requests []isc.Request) []gpa.MessageOut {
	decidedBaseAnchor := aggregatedProposals.DecidedBaseAnchor()
	stateAnchor := isc.NewStateAnchor(decidedBaseAnchor.Anchor(), decidedBaseAnchor.ISCPackage())
	gasCoins := aggregatedProposals.AggregatedGasCoins()
	// FIXME we need only one
	if len(gasCoins) != 1 {
		panic("FIXME we support only one gas coin now")
	}
	gasCoin := gasCoins[0]

	c.output.NeedVMResult = &vm.VMTask{
		Processors:           c.processorCache,
		Anchor:               &stateAnchor,
		GasCoin:              gasCoin,
		L1Params:             aggregatedProposals.AggregatedL1Params(),
		Store:                c.chainStore,
		Requests:             aggregatedProposals.OrderedRequests(requests, *randomness),
		Timestamp:            aggregatedProposals.AggregatedTime(),
		Entropy:              *randomness,
		ValidatorFeeTarget:   aggregatedProposals.ValidatorFeeTarget(*randomness),
		EstimateGasMode:      false,
		EnableGasBurnLogging: false,
		Log:                  c.log.NewChildLogger("VM"),
		Migrations:           allmigrations.DefaultScheme,
	}
	return c.subTX.AnchorDecided(decidedBaseAnchor)
}

func (c *Consensus) uponVMOutputReceived(vmResult *vm.VMTaskResult, aggregatedProposals *batchproposal.AggregatedBatchProposals) []gpa.MessageOut {
	c.output.NeedVMResult = nil
	if len(vmResult.RequestResults) == 0 {
		// No requests were processed, don't have what to do.
		// Will need to retry the consensus with the next log index some time later.
		c.log.LogInfof("Terminating consensus with status=Skipped, 0 requests processed.")
		c.output.Status = Skipped
		c.term.haveOutputProduced()
		return nil
	}

	// Make sure all the fields in the TX are ordered properly.
	unsignedTX := vmResult.UnsignedTransaction
	txData := c.makeTransactionData(&unsignedTX, aggregatedProposals)
	txBytes := c.makeTransactionSigningBytes(txData)
	c.log.LogDebugf("VM produced TxDataBytes=%s", hex.EncodeToString(c.makeTransactionDataBytes(txData)))
	return slices.Concat(
		c.subStateMgr.BlockProduced(vmResult.StateDraft),
		c.subTX.UnsignedTXReceived(txData),
		c.subDistributedSignature.MessageToSignReceived(txBytes),
	)
}

////////////////////////////////////////////////////////////////////////////////
// TX

func (c *Consensus) makeTransactionData(pt *iotago.ProgrammableTransaction, aggregatedProposals *batchproposal.AggregatedBatchProposals) *iotago.TransactionData {
	sender := c.dkShare.GetAddress().AsIotaAddress()
	l1params := aggregatedProposals.AggregatedL1Params()
	gasPrice := l1params.Protocol.ReferenceGasPrice.Uint64()
	gasBudget := pt.EstimateGasBudget(gasPrice)
	gasPaymentCoinRef := aggregatedProposals.AggregatedGasCoins()
	gasPayment := make([]*iotago.ObjectRef, len(gasPaymentCoinRef))
	for i, coinRef := range gasPaymentCoinRef {
		gasPayment[i] = coinRef.Ref
	}

	tx := iotago.NewProgrammable(sender, *pt, gasPayment, gasBudget, gasPrice)
	return &tx
}

func (c *Consensus) makeTransactionDataBytes(txData *iotago.TransactionData) []byte {
	txnBytes, err := bcs.Marshal(txData)
	if err != nil {
		panic(fmt.Errorf("uponVMOutputReceived: cannot serialize the tx: %w", err))
	}
	return txnBytes
}

func (c *Consensus) makeTransactionSigningBytes(txData *iotago.TransactionData) []byte {
	txnBytes := c.makeTransactionDataBytes(txData)
	txnBytes = iotasigner.MessageWithIntent(iotasigner.DefaultIntent(), txnBytes)
	txnBytesHash := blake2b.Sum256(txnBytes)
	return txnBytesHash[:]
}

// Everything is ready for the output TX, produce it.
func (c *Consensus) uponTXInputsReady(decidedAnchor *isc.StateAnchor, unsignedTX *iotago.TransactionData, block state.Block, signature []byte) []gpa.MessageOut {
	suiSignature := cryptolib.NewSignature(c.dkShare.GetSharedPublic(), signature).AsIotaSignature()
	signedTX := iotasigner.NewSignedTransaction(unsignedTX, suiSignature)
	c.output.Result = &Result{
		DecidedAnchor: decidedAnchor,
		Transaction:   signedTX,
		Block:         block,
	}
	c.output.Status = Completed
	c.log.LogInfof("Terminating consensus with status=Completed")
	c.term.haveOutputProduced()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TERM

func (c *Consensus) uponTerminationCondition() {
	c.output.Terminated = true
}
