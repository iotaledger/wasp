// Consensus. A single instance of it.
//
// We move all the synchronization logic to separate objects (upon_...). They are
// responsible for waiting specific data and then triggering the next state action
// once. This way we hope to solve a lot of race conditions gracefully. The `upon`
// predicates and the corresponding done functions should not depend on each other.
// If some data is needed at several places, it should be passed to several predicates.
package cons

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/minio/blake2b-simd"
	"github.com/samber/lo"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/chain/cons/bp"
	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acs"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

type Cons interface {
	AsGPA() gpa.GPA
}

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
	DecidedAO   *isc.StateAnchor              // The consumed state anchor.
	Transaction *iotasigner.SignedTransaction // The TX for committing the block.
	Block       state.Block                   // The state diff produced.
}

func (r *Result) String() string {
	return fmt.Sprintf(
		"{cons.Result, txDigest=%s, baseAO=%v, outBlockHash=%v}",
		lo.Must(r.Transaction.Digest()),
		r.DecidedAO,
		r.Block.Hash(),
	)
}

type consImpl struct {
	chainID          isc.ChainID
	chainStore       state.Store
	edSuite          suites.Suite    // For signatures.
	blsSuite         suites.Suite    // For randomness only.
	dkShare          tcrypto.DKShare // The current committee's keys.
	rotateTo         *iotago.Address // If non-nil and differs from the dkShare, then rotation is suggested.
	processorCache   *processors.Config
	nodeIDs          []gpa.NodeID
	me               gpa.NodeID
	f                int
	asGPA            gpa.GPA
	dss              dss.DSS
	acs              acs.ACS
	subMP            SyncMP         // Mempool.
	subSM            SyncSM         // StateMgr.
	subNC            SyncNC         // Synchronization with the NodeConn.
	subDSS           SyncDSS        // Distributed Schnorr Signature.
	subACS           SyncACS        // Asynchronous Common Subset.
	subRND           SyncRND        // Randomness.
	subVM            SyncVM         // Virtual Machine.
	subTX            SyncTX         // Building final TX.
	term             *termCondition // To detect, when this instance can be terminated.
	msgWrapper       *gpa.MsgWrapper
	output           *Output
	validatorAgentID isc.AgentID
	log              log.Logger
}

const (
	subsystemTypeDSS byte = iota
	subsystemTypeACS
)

var (
	_ gpa.GPA = &consImpl{}
	_ Cons    = &consImpl{}
)

func New( //nolint:funlen
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
) Cons {
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
	acsCCInstFunc := func(nodeID gpa.NodeID, round int) gpa.GPA {
		var roundBin [4]byte
		binary.BigEndian.PutUint32(roundBin[:], uint32(round))
		sid := hashing.HashDataBlake2b(instID, nodeID[:], roundBin[:]).Bytes()
		realCC := blssig.New(blsSuite, nodeIDs, dkShare.BLSCommits(), dkShare.BLSPriShare(), int(dkShare.BLSThreshold()), me, sid, acsLog)
		return semi.New(round, realCC)
	}
	c := &consImpl{
		chainID:          chainID,
		chainStore:       chainStore,
		edSuite:          edSuite,
		blsSuite:         blsSuite,
		dkShare:          dkShare,
		rotateTo:         rotateTo,
		processorCache:   processorCache,
		nodeIDs:          nodeIDs,
		me:               me,
		f:                f,
		dss:              dss.New(edSuite, nodeIDs, nodePKs, f, me, myKyberKeys.Private, longTermDKS, log.NewChildLogger("DSS")),
		acs:              acs.New(nodeIDs, me, f, acsCCInstFunc, acsLog),
		output:           &Output{Status: Running},
		log:              log,
		validatorAgentID: validatorAgentID,
	}
	c.asGPA = gpa.NewOwnHandler(me, c)
	c.msgWrapper = gpa.NewMsgWrapper(msgTypeWrapped, c.msgWrapperFunc)
	c.subMP = NewSyncMP(
		c.uponMPProposalInputsReady,
		c.uponMPProposalReceived,
		c.uponMPRequestsNeeded,
		c.uponMPRequestsReceived,
	)
	c.subSM = NewSyncSM(
		c.uponSMStateProposalQueryInputsReady,
		c.uponSMStateProposalReceived,
		c.uponSMDecidedStateQueryInputsReady,
		c.uponSMDecidedStateReceived,
		c.uponSMSaveProducedBlockInputsReady,
		c.uponSMSaveProducedBlockDone,
	)
	c.subNC = NewSyncNC(
		c.uponNCInputsReady,
		c.uponNCOutputReady,
	)
	c.subDSS = NewSyncDSS(
		c.uponDSSInitialInputsReady,
		c.uponDSSIndexProposalReady,
		c.uponDSSSigningInputsReceived,
		c.uponDSSOutputReady,
	)
	c.subACS = NewSyncACS(
		c.uponACSInputsReceived,
		c.uponACSOutputReceived,
		c.uponACSTerminated,
	)
	c.subRND = NewSyncRND(
		int(dkShare.BLSThreshold()),
		c.uponRNDInputsReady,
		c.uponRNDSigSharesReady,
	)
	c.subVM = NewSyncVM(
		c.uponVMInputsReceived,
		c.uponVMOutputReceived,
	)
	c.subTX = NewSyncTX(
		c.uponTXInputsReady,
	)
	c.term = newTermCondition(
		c.uponTerminationCondition,
	)
	return c
}

// Used to select a target subsystem for a wrapped message received.
func (c *consImpl) msgWrapperFunc(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == subsystemTypeDSS {
		if index != 0 {
			return nil, fmt.Errorf("unexpected DSS index: %v", index)
		}
		return c.dss.AsGPA(), nil
	}
	if subsystem == subsystemTypeACS {
		if index != 0 {
			return nil, fmt.Errorf("unexpected ACS index: %v", index)
		}
		return c.acs.AsGPA(), nil
	}
	return nil, fmt.Errorf("unexpected subsystem: %v", subsystem)
}

func (c *consImpl) AsGPA() gpa.GPA {
	return c.asGPA
}

func (c *consImpl) Input(input gpa.Input) gpa.OutMessages {
	switch input := input.(type) {
	case *inputTimeData:
		// ignore this to filter out ridiculously excessive logging
	default:
		c.log.LogDebugf("Input %T: %+v", input, input)
	}

	switch input := input.(type) {
	case *inputProposal:
		c.log.LogInfof("Consensus started, received %v", input.String())
		return gpa.NoMessages().
			AddAll(c.subNC.HaveInputAnchor(input.baseAliasOutput)).
			AddAll(c.subMP.BaseAliasOutputReceived(input.baseAliasOutput)).
			AddAll(c.subSM.ProposedBaseAliasOutputReceived(input.baseAliasOutput)).
			AddAll(c.subDSS.InitialInputReceived())
	case *inputRotateTo:
		// We can update the rotation address while consensus is running.
		// New value will be used, if decision has not been made yet.
		c.rotateTo = input.address
		return nil
	case *inputMempoolProposal:
		return c.subMP.ProposalReceived(input.requestRefs)
	case *inputMempoolRequests:
		return c.subMP.RequestsReceived(input.requests)
	case *inputStateMgrProposalConfirmed:
		return c.subSM.StateProposalConfirmedByStateMgr()
	case *inputStateMgrDecidedVirtualState:
		return c.subSM.DecidedVirtualStateReceived(input.chainState)
	case *inputStateMgrBlockSaved:
		return c.subSM.BlockSaved(input.block)
	case *inputTimeData:
		return c.subACS.TimeDataReceived(input.timeData)
	case *inputL1Info:
		return c.subNC.HaveL1Info(input.gasCoins, input.l1params)
	case *inputVMResult:
		return c.subVM.VMResultReceived(input.task)
	}
	panic(fmt.Errorf("unexpected input: %v", input))
}

// Implements the gpa.GPA interface.
// Here we route all the messages.
func (c *consImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msgT := msg.(type) {
	case *msgBLSPartialSig:
		return c.subRND.BLSPartialSigReceived(msgT.Sender(), msgT.partialSig)
	case *gpa.WrappingMsg:
		sub, subMsgs, err := c.msgWrapper.DelegateMessage(msgT)
		if err != nil {
			c.log.LogWarnf("unexpected wrapped message: %w", err)
			return nil
		}
		msgs := gpa.NoMessages().AddAll(subMsgs)
		switch msgT.Subsystem() {
		case subsystemTypeACS:
			return msgs.AddAll(c.subACS.ACSOutputReceived(sub.Output()))
		case subsystemTypeDSS:
			return msgs.AddAll(c.subDSS.DSSOutputReceived(sub.Output()))
		default:
			c.log.LogWarnf("unexpected subsystem after check: %+v", msg)
			return nil
		}
	}
	panic(fmt.Errorf("unexpected message: %v", msg))
}

func (c *consImpl) Output() gpa.Output {
	return c.output // Always non-nil.
}

func (c *consImpl) StatusString() string {
	// We con't include RND here, maybe that's less important, and visible from the VM status.
	return fmt.Sprintf("{consImpl⟨%v⟩,%v,%v,%v,%v,%v,%v,%v}",
		c.output.Status,
		c.subSM.String(),
		c.subMP.String(),
		c.subNC.String(),
		c.subDSS.String(),
		c.subACS.String(),
		c.subVM.String(),
		c.subTX.String(),
	)
}

////////////////////////////////////////////////////////////////////////////////
// MP -- MemPool

func (c *consImpl) uponMPProposalInputsReady(baseAliasOutput *isc.StateAnchor) gpa.OutMessages {
	if baseAliasOutput == nil {
		// If the base AO is nil, we are not going to propose any requests.
		return c.subMP.ProposalReceived([]*isc.RequestRef{})
	}
	c.output.NeedMempoolProposal = baseAliasOutput
	return nil
}

func (c *consImpl) uponMPProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages {
	c.output.NeedMempoolProposal = nil
	msgs := gpa.NoMessages()
	msgs.AddAll(c.subACS.MempoolRequestsReceived(requestRefs))
	msgs.AddAll(c.subNC.HaveRequests())
	return msgs
}

func (c *consImpl) uponMPRequestsNeeded(requestRefs []*isc.RequestRef) gpa.OutMessages {
	c.output.NeedMempoolRequests = requestRefs
	return nil
}

func (c *consImpl) uponMPRequestsReceived(requests []isc.Request) gpa.OutMessages {
	c.output.NeedMempoolRequests = nil
	return c.subVM.RequestsReceived(requests)
}

////////////////////////////////////////////////////////////////////////////////
// SM -- StateManager

func (c *consImpl) uponSMStateProposalQueryInputsReady(baseAliasOutput *isc.StateAnchor) gpa.OutMessages {
	if baseAliasOutput == nil {
		// Don't wait for the state if no base AO is known.
		return c.subSM.StateProposalConfirmedByStateMgr()
	}
	c.output.NeedStateMgrStateProposal = baseAliasOutput
	return nil
}

func (c *consImpl) uponSMStateProposalReceived(proposedAliasOutput *isc.StateAnchor) gpa.OutMessages {
	c.output.NeedStateMgrStateProposal = nil
	msgs := gpa.NoMessages()
	msgs.AddAll(c.subACS.StateProposalReceived(proposedAliasOutput))
	msgs.AddAll(c.subNC.HaveState())
	return msgs
}

func (c *consImpl) uponSMDecidedStateQueryInputsReady(decidedBaseAliasOutput *isc.StateAnchor) gpa.OutMessages {
	c.output.NeedStateMgrDecidedState = decidedBaseAliasOutput
	return nil
}

func (c *consImpl) uponSMDecidedStateReceived(chainState state.State) gpa.OutMessages {
	c.output.NeedStateMgrDecidedState = nil
	return c.subVM.DecidedStateReceived(chainState)
}

func (c *consImpl) uponSMSaveProducedBlockInputsReady(producedBlock state.StateDraft) gpa.OutMessages {
	if producedBlock == nil {
		// Don't have a block to save in the case of self-governed rotation.
		// So mark it as saved immediately.
		return c.subSM.BlockSaved(nil)
	}
	c.output.NeedStateMgrSaveBlock = producedBlock
	return nil
}

func (c *consImpl) uponSMSaveProducedBlockDone(block state.Block) gpa.OutMessages {
	c.output.NeedStateMgrSaveBlock = nil
	return c.subTX.BlockSaved(block)
}

////////////////////////////////////////////////////////////////////////////////
// NC

func (c *consImpl) uponNCInputsReady(anchor *isc.StateAnchor) gpa.OutMessages {
	if anchor == nil {
		c.log.LogDebugf("ACS got ⊥ as input, no L1 info can be fetched.")
		return c.subACS.L1InfoReceived([]*coin.CoinWithRef{}, nil)
	}
	c.output.NeedNodeConnL1Info = anchor
	return nil
}

func (c *consImpl) uponNCOutputReady(gasCoins []*coin.CoinWithRef, l1params *parameters.L1Params) gpa.OutMessages {
	c.log.LogDebugf("L1 info received, gasCoins=%v, l1Params=%v", gasCoins, l1params)
	c.output.NeedNodeConnL1Info = nil
	return c.subACS.L1InfoReceived(gasCoins, l1params)
}

////////////////////////////////////////////////////////////////////////////////
// DSS

func (c *consImpl) uponDSSInitialInputsReady() gpa.OutMessages {
	c.log.LogDebugf("uponDSSInitialInputsReady")
	sub, subMsgs, err := c.msgWrapper.DelegateInput(subsystemTypeDSS, 0, dss.NewInputStart())
	if err != nil {
		panic(fmt.Errorf("cannot provide input to DSS: %w", err))
	}
	return gpa.NoMessages().
		AddAll(subMsgs).
		AddAll(c.subDSS.DSSOutputReceived(sub.Output()))
}

func (c *consImpl) uponDSSIndexProposalReady(indexProposal []int) gpa.OutMessages {
	c.log.LogDebugf("uponDSSIndexProposalReady")
	return c.subACS.DSSIndexProposalReceived(indexProposal)
}

func (c *consImpl) uponDSSSigningInputsReceived(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.OutMessages {
	c.log.LogDebugf("uponDSSSigningInputsReceived(decidedIndexProposals=%+v, H(messageToSign)=%v)", decidedIndexProposals, hashing.HashDataBlake2b(messageToSign))
	dssDecidedInput := dss.NewInputDecided(decidedIndexProposals, messageToSign)
	subDSS, subMsgs, err := c.msgWrapper.DelegateInput(subsystemTypeDSS, 0, dssDecidedInput)
	if err != nil {
		panic(fmt.Errorf("cannot provide inputs for signing: %w", err))
	}
	return gpa.NoMessages().
		AddAll(subMsgs).
		AddAll(c.subDSS.DSSOutputReceived(subDSS.Output()))
}

func (c *consImpl) uponDSSOutputReady(signature []byte) gpa.OutMessages {
	c.log.LogDebugf("uponDSSOutputReady")
	return c.subTX.SignatureReceived(signature)
}

////////////////////////////////////////////////////////////////////////////////
// ACS

func (c *consImpl) uponACSInputsReceived(
	baseAliasOutput *isc.StateAnchor, // Can be nil.
	requestRefs []*isc.RequestRef,
	dssIndexProposal []int,
	timeData time.Time,
	gasCoins []*coin.CoinWithRef, // Can be nil.
	l1params *parameters.L1Params, // Can be nil.
) gpa.OutMessages {
	rotateTo := c.rotateTo
	if rotateTo != nil && rotateTo.Equals(*c.dkShare.GetAddress().AsIotaAddress()) {
		// Do not propose to rotate to the existing committee.
		rotateTo = nil
	}
	batchProposal := bp.NewBatchProposal(
		*c.dkShare.GetIndex(),
		baseAliasOutput, // Will be NIL in the case of ⊥ proposal.
		util.NewFixedSizeBitVector(c.dkShare.GetN()).SetBits(dssIndexProposal),
		rotateTo,
		timeData,
		c.validatorAgentID,
		requestRefs, // Will be [] in the case of ⊥ proposal.
		gasCoins,    // Will be NIL in the case of ⊥ proposal.
		l1params,    // Will be NIL in the case of ⊥ proposal.
	)
	subACS, subMsgs, err := c.msgWrapper.DelegateInput(subsystemTypeACS, 0, batchProposal.Bytes())
	if err != nil {
		panic(fmt.Errorf("cannot provide input to the ACS: %w", err))
	}
	return gpa.NoMessages().
		AddAll(subMsgs).
		AddAll(c.subACS.ACSOutputReceived(subACS.Output()))
}

func (c *consImpl) uponACSOutputReceived(outputValues map[gpa.NodeID][]byte) gpa.OutMessages {
	aggr := bp.AggregateBatchProposals(outputValues, c.nodeIDs, c.f, c.log)
	if aggr.ShouldBeSkipped() {
		// Cannot proceed with such proposals.
		// Have to retry the consensus after some time with the next log index.
		c.log.LogInfof("Terminating consensus with status=Skipped, there is no way to aggregate batch proposal.")
		c.output.Status = Skipped
		c.term.haveOutputProduced()
		return nil
	}
	bao := aggr.DecidedBaseAliasOutput()
	baoID := bao.GetObjectID()
	reqs := aggr.DecidedRequestRefs()
	c.log.LogDebugf("ACS decision: baseAO=%v, requests=%v", bao, reqs)
	if aggr.DecidedRotateTo() != nil {
		c.log.LogDebugf("Will rotate to %v", aggr.DecidedRotateTo().ToHex())
		rotationPTB := vmtxbuilder.NewAnchorTransactionBuilder(bao.ISCPackage(), bao, c.dkShare.GetAddress())
		rotationPTB.RotationTransaction(aggr.DecidedRotateTo())
		rotationPTX := rotationPTB.BuildTransactionEssence(bao.GetStateMetadata(), 0)
		rotationTXD := c.makeTransactionData(&rotationPTX, aggr)
		rotationTXB := c.makeTransactionSigningBytes(rotationTXD)
		c.log.LogDebugf("Rotation TxDataBytes=%s", hex.EncodeToString(c.makeTransactionDataBytes(rotationTXD)))
		return gpa.NoMessages().
			AddAll(c.subTX.UnsignedTXReceived(rotationTXD)).
			AddAll(c.subTX.BlockSaved(nil)).
			AddAll(c.subTX.AnchorDecided(bao)).
			AddAll(c.subDSS.MessageToSignReceived(rotationTXB)).
			AddAll(c.subDSS.DecidedIndexProposalsReceived(aggr.DecidedDSSIndexProposals()))
	}
	return gpa.NoMessages().
		AddAll(c.subMP.RequestsNeeded(reqs)).
		AddAll(c.subSM.DecidedVirtualStateNeeded(bao)).
		AddAll(c.subVM.DecidedBatchProposalsReceived(aggr)).
		AddAll(c.subRND.CanProceed(baoID.Bytes())).
		AddAll(c.subDSS.DecidedIndexProposalsReceived(aggr.DecidedDSSIndexProposals()))
}

func (c *consImpl) uponACSTerminated() {
	c.term.haveAcsTerminated()
}

////////////////////////////////////////////////////////////////////////////////
// RND

func (c *consImpl) uponRNDInputsReady(dataToSign []byte) gpa.OutMessages {
	sigShare, err := c.dkShare.BLSSignShare(dataToSign)
	if err != nil {
		panic(fmt.Errorf("cannot sign share for randomness: %w", err))
	}
	msgs := gpa.NoMessages()
	for _, nid := range c.nodeIDs {
		msgs.Add(newMsgBLSPartialSig(c.blsSuite, nid, sigShare))
	}
	return msgs
}

func (c *consImpl) uponRNDSigSharesReady(dataToSign []byte, partialSigs map[gpa.NodeID][]byte) (bool, gpa.OutMessages) {
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

func (c *consImpl) uponVMInputsReceived(aggregatedProposals *bp.AggregatedBatchProposals, chainState state.State, randomness *hashing.HashValue, requests []isc.Request) gpa.OutMessages {
	decidedBaseAliasOutput := aggregatedProposals.DecidedBaseAliasOutput()
	stateAnchor := isc.NewStateAnchor(decidedBaseAliasOutput.Anchor(), decidedBaseAliasOutput.ISCPackage())
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
	return c.subTX.AnchorDecided(decidedBaseAliasOutput)
}

func (c *consImpl) uponVMOutputReceived(vmResult *vm.VMTaskResult, aggregatedProposals *bp.AggregatedBatchProposals) gpa.OutMessages {
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
	return gpa.NoMessages().
		AddAll(c.subSM.BlockProduced(vmResult.StateDraft)).
		AddAll(c.subTX.UnsignedTXReceived(txData)).
		AddAll(c.subDSS.MessageToSignReceived(txBytes))
}

////////////////////////////////////////////////////////////////////////////////
// TX

func (c *consImpl) makeTransactionData(pt *iotago.ProgrammableTransaction, aggregatedProposals *bp.AggregatedBatchProposals) *iotago.TransactionData {
	var sender *iotago.Address = c.dkShare.GetAddress().AsIotaAddress()
	var l1params *parameters.L1Params = aggregatedProposals.AggregatedL1Params()
	gasPrice := l1params.Protocol.ReferenceGasPrice.Uint64()
	var gasBudget uint64 = pt.EstimateGasBudget(gasPrice)
	var gasPaymentCoinRef []*coin.CoinWithRef = aggregatedProposals.AggregatedGasCoins()
	gasPayment := make([]*iotago.ObjectRef, len(gasPaymentCoinRef))
	for i, coinRef := range gasPaymentCoinRef {
		gasPayment[i] = coinRef.Ref
	}

	tx := iotago.NewProgrammable(sender, *pt, gasPayment, gasBudget, gasPrice)
	return &tx
}

func (c *consImpl) makeTransactionDataBytes(txData *iotago.TransactionData) []byte {
	txnBytes, err := bcs.Marshal(txData)
	if err != nil {
		panic(fmt.Errorf("uponVMOutputReceived: cannot serialize the tx: %w", err))
	}
	return txnBytes
}

func (c *consImpl) makeTransactionSigningBytes(txData *iotago.TransactionData) []byte {
	txnBytes := c.makeTransactionDataBytes(txData)
	txnBytes = iotasigner.MessageWithIntent(iotasigner.DefaultIntent(), txnBytes)
	txnBytesHash := blake2b.Sum256(txnBytes)
	return txnBytesHash[:]
}

// Everything is ready for the output TX, produce it.
func (c *consImpl) uponTXInputsReady(decidedAO *isc.StateAnchor, unsignedTX *iotago.TransactionData, block state.Block, signature []byte) gpa.OutMessages {
	suiSignature := cryptolib.NewSignature(c.dkShare.GetSharedPublic(), signature).AsIotaSignature()
	signedTX := iotasigner.NewSignedTransaction(unsignedTX, suiSignature)
	c.output.Result = &Result{
		DecidedAO:   decidedAO,
		Transaction: signedTX,
		Block:       block,
	}
	c.output.Status = Completed
	c.log.LogInfof("Terminating consensus with status=Completed")
	c.term.haveOutputProduced()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TERM

func (c *consImpl) uponTerminationCondition() {
	c.output.Terminated = true
}
