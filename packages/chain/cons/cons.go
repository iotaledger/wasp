// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Consensus. Single instance of it.
//
// Used sub-protocols (on the same thread):
//   - DSS -- Distributed Schnorr Signature
//   - ACS -- Asynchronous Common Subset
//
// Used components (running on other threads):
//   - Mempool
//   - StateMgr
//   - VM
//
// > INPUT: baseAliasOutputID
// > ON Startup:
// >     Start a DSS.
// >     Ask Mempool for backlog (based on baseAliasOutputID).
// >     Ask StateMgr for a virtual state (based on baseAliasOutputID).
// > UPON Reception of responses from Mempool, StateMgr and DSS NonceIndexes:
// >     Produce a batch proposal.
// >     Start the ACS.
// > UPON Reception of ACS output:
// >     IF result is possible THEN
// >         Submit agreed NonceIndexes to DSS.
// >         Send the BLS partial signature.
// >     ELSE
// >         OUTPUT SKIP
// > UPON Reception of N-2F BLS partial signatures:
// >     Start VM.
// > UPON Reception of VM Result:
// >     IF result is non-empty THEN
// >         Save the produced block to SM.
// >         Submit the result hash to the DSS.
// >     ELSE
// >         OUTPUT SKIP
// > UPON Reception of VM Result and a signature from the DSS
// >     IF rotation THEN
// >        OUTPUT Signed Governance TX.
// >     ELSE
// >        Save the block to the StateMgr.
// >        OUTPUT Signed State Transition TX
//
// We move all the synchronization logic to separate objects (upon_...). They are
// responsible for waiting specific data and then triggering the next state action
// once. This way we hope to solve a lot of race conditions gracefully. The `upon`
// predicates and the corresponding done functions should not depend on each other.
// If some data is needed at several places, it should be passed to several predicates.
//
// TODO: Handle the requests gracefully in the VM before getting the initTX.
// TODO: Reconsider the termination. Do we need to wait for DSS, RND?
package cons

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"time"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/hive.go/core/identity"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cons/bp"
	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acs"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/rotate"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type Cons interface {
	AsGPA() gpa.GPA
}

type OutputStatus byte

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
	NeedMempoolProposal       *isc.AliasOutputWithID // Requests for the mempool are needed for this Base Alias Output.
	NeedMempoolRequests       []*isc.RequestRef      // Request payloads are needed from mempool for this IDs/Hash.
	NeedStateMgrStateProposal *isc.AliasOutputWithID // Query for a proposal for Virtual State (it will go to the batch proposal).
	NeedStateMgrDecidedState  *isc.AliasOutputWithID // Query for a decided Virtual State to be used by VM.
	NeedStateMgrSaveBlock     state.StateDraft       // Ask StateMgr to save the produced block.
	NeedVMResult              *vm.VMTask             // VM Result is needed for this (agreed) batch.
	//
	// Following is the final result.
	// All the fields are filled, if State == Completed.
	ResultTransaction     *iotago.Transaction
	ResultNextAliasOutput *isc.AliasOutputWithID
	ResultState           state.StateDraft
}

type consImpl struct {
	chainID        isc.ChainID
	chainStore     state.Store
	edSuite        suites.Suite // For signatures.
	blsSuite       suites.Suite // For randomness only.
	dkShare        tcrypto.DKShare
	processorCache *processors.Cache
	nodeIDs        []gpa.NodeID
	me             gpa.NodeID
	f              int
	asGPA          gpa.GPA
	dss            dss.DSS
	acs            acs.ACS
	subMP          SyncMP         // Mempool.
	subSM          SyncSM         // StateMgr.
	subDSS         SyncDSS        // Distributed Schnorr Signature.
	subACS         SyncACS        // Asynchronous Common Subset.
	subRND         SyncRND        // Randomness.
	subVM          SyncVM         // Virtual Machine.
	subTX          SyncTX         // Building final TX.
	term           *termCondition // To detect, when this instance can be terminated.
	msgWrapper     *gpa.MsgWrapper
	output         *Output
	log            *logger.Logger
}

const (
	subsystemTypeDSS byte = iota
	subsystemTypeACS
)

var (
	_ gpa.GPA = &consImpl{}
	_ Cons    = &consImpl{}
)

func New(
	chainID isc.ChainID,
	chainStore state.Store,
	me gpa.NodeID,
	mySK *cryptolib.PrivateKey,
	dkShare tcrypto.DKShare,
	processorCache *processors.Cache,
	instID []byte,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	log *logger.Logger,
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
	longTermDKS := dkShare.DSSSecretShare()
	acsCCInstFunc := func(nodeID gpa.NodeID, round int) gpa.GPA {
		var roundBin [4]byte
		binary.BigEndian.PutUint32(roundBin[:], uint32(round))
		sid := hashing.HashDataBlake2b(instID, nodeID[:], roundBin[:]).Bytes()
		realCC := blssig.New(blsSuite, nodeIDs, dkShare.BLSCommits(), dkShare.BLSPriShare(), int(dkShare.BLSThreshold()), me, sid, log)
		return semi.New(round, realCC)
	}
	c := &consImpl{
		chainID:        chainID,
		chainStore:     chainStore,
		edSuite:        edSuite,
		blsSuite:       blsSuite,
		dkShare:        dkShare,
		processorCache: processorCache,
		nodeIDs:        nodeIDs,
		me:             me,
		f:              f,
		dss:            dss.New(edSuite, nodeIDs, nodePKs, f, me, myKyberKeys.Private, longTermDKS, log),
		acs:            acs.New(nodeIDs, me, f, acsCCInstFunc, log),
		output:         &Output{Status: Running},
		log:            log,
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
	case *inputProposal:
		c.log.Debugf("received %v", input.String())
		return gpa.NoMessages().
			AddAll(c.subMP.BaseAliasOutputReceived(input.baseAliasOutput)).
			AddAll(c.subSM.ProposedBaseAliasOutputReceived(input.baseAliasOutput)).
			AddAll(c.subDSS.InitialInputReceived())
	case *inputMempoolProposal:
		return c.subMP.ProposalReceived(input.requestRefs)
	case *inputMempoolRequests:
		return c.subMP.RequestsReceived(input.requests)
	case *inputStateMgrProposalConfirmed:
		return c.subSM.StateProposalConfirmedByStateMgr()
	case *inputStateMgrDecidedVirtualState:
		return c.subSM.DecidedVirtualStateReceived(input.chainState)
	case *inputStateMgrBlockSaved:
		return c.subSM.BlockSaved()
	case *inputTimeData:
		return c.subACS.TimeDataReceived(input.timeData)
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
		return c.subRND.BLSPartialSigReceived(msgT.sender, msgT.partialSig)
	case *gpa.WrappingMsg:
		sub, subMsgs, err := c.msgWrapper.DelegateMessage(msgT)
		if err != nil {
			c.log.Warnf("unexpected wrapped message: %w", err)
			return nil
		}
		msgs := gpa.NoMessages().AddAll(subMsgs)
		switch msgT.Subsystem() {
		case subsystemTypeACS:
			return msgs.AddAll(c.subACS.ACSOutputReceived(sub.Output()))
		case subsystemTypeDSS:
			return msgs.AddAll(c.subDSS.DSSOutputReceived(sub.Output()))
		}
		panic(fmt.Errorf("unexpected subsystem after check: %+v", msg))
	}
	panic(fmt.Errorf("unexpected message: %v", msg))
}

func (c *consImpl) Output() gpa.Output {
	return c.output // Always non-nil.
}

func (c *consImpl) StatusString() string {
	// We con't include RND here, maybe that's less important, and visible from the VM status.
	return fmt.Sprintf("{consImpl,%v,%v,%v,%v,%v,%v}",
		c.subSM.String(),
		c.subMP.String(),
		c.subDSS.String(),
		c.subACS.String(),
		c.subVM.String(),
		c.subTX.String(),
	)
}

////////////////////////////////////////////////////////////////////////////////
// MP -- MemPool

func (c *consImpl) uponMPProposalInputsReady(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	c.output.NeedMempoolProposal = baseAliasOutput
	return nil
}

func (c *consImpl) uponMPProposalReceived(requestRefs []*isc.RequestRef) gpa.OutMessages {
	c.output.NeedMempoolProposal = nil
	return c.subACS.MempoolRequestsReceived(requestRefs)
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

func (c *consImpl) uponSMStateProposalQueryInputsReady(baseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	c.output.NeedStateMgrStateProposal = baseAliasOutput
	return nil
}

func (c *consImpl) uponSMStateProposalReceived(proposedAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
	c.output.NeedStateMgrStateProposal = nil
	return c.subACS.StateProposalReceived(proposedAliasOutput)
}

func (c *consImpl) uponSMDecidedStateQueryInputsReady(decidedBaseAliasOutput *isc.AliasOutputWithID) gpa.OutMessages {
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
		return c.subSM.BlockSaved()
	}
	c.output.NeedStateMgrSaveBlock = producedBlock
	return nil
}

func (c *consImpl) uponSMSaveProducedBlockDone() gpa.OutMessages {
	c.output.NeedStateMgrSaveBlock = nil
	return c.subTX.BlockSaved()
}

////////////////////////////////////////////////////////////////////////////////
// DSS

func (c *consImpl) uponDSSInitialInputsReady() gpa.OutMessages {
	sub, subMsgs, err := c.msgWrapper.DelegateInput(subsystemTypeDSS, 0, dss.NewInputStart())
	if err != nil {
		panic(fmt.Errorf("cannot provide input to DSS: %w", err))
	}
	return gpa.NoMessages().
		AddAll(subMsgs).
		AddAll(c.subDSS.DSSOutputReceived(sub.Output()))
}

func (c *consImpl) uponDSSIndexProposalReady(indexProposal []int) gpa.OutMessages {
	return c.subACS.DSSIndexProposalReceived(indexProposal)
}

func (c *consImpl) uponDSSSigningInputsReceived(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.OutMessages {
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
	return c.subTX.SignatureReceived(signature)
}

////////////////////////////////////////////////////////////////////////////////
// ACS

func (c *consImpl) uponACSInputsReceived(baseAliasOutput *isc.AliasOutputWithID, requestRefs []*isc.RequestRef, dssIndexProposal []int, timeData time.Time) gpa.OutMessages {
	batchProposal := bp.NewBatchProposal(
		*c.dkShare.GetIndex(),
		baseAliasOutput,
		util.NewFixedSizeBitVector(int(c.dkShare.GetN())).SetBits(dssIndexProposal),
		timeData,
		isc.NewContractAgentID(c.chainID, 0),
		requestRefs,
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
		c.output.Status = Skipped
		c.term.haveOutputProduced()
		return nil
	}
	bao := aggr.DecidedBaseAliasOutput()
	baoID := bao.OutputID()
	return gpa.NoMessages().
		AddAll(c.subMP.RequestsNeeded(aggr.DecidedRequestRefs())).
		AddAll(c.subSM.DecidedVirtualStateNeeded(bao)).
		AddAll(c.subVM.DecidedBatchProposalsReceived(aggr)).
		AddAll(c.subRND.CanProceed(baoID[:])).
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
		c.log.Warnf("Cannot reconstruct BLS signature from %v/%v sigShares: %v", len(partialSigs), c.dkShare.GetN(), err)
		return false, nil // Continue to wait for other sig shares.
	}
	return true, c.subVM.RandomnessReceived(hashing.HashDataBlake2b(sig.Signature.Bytes()))
}

////////////////////////////////////////////////////////////////////////////////
// VM

func (c *consImpl) uponVMInputsReceived(aggregatedProposals *bp.AggregatedBatchProposals, chainState state.State, randomness *hashing.HashValue, requests []isc.Request) gpa.OutMessages {
	// TODO: chainState state.State is not used for now. That's because VM takes it form the store by itself.
	// The decided base alias output can be different from that we have proposed!
	decidedBaseAliasOutput := aggregatedProposals.DecidedBaseAliasOutput()
	c.output.NeedVMResult = &vm.VMTask{
		Processors:             c.processorCache,
		AnchorOutput:           decidedBaseAliasOutput.GetAliasOutput(),
		AnchorOutputID:         decidedBaseAliasOutput.OutputID(),
		Store:                  c.chainStore,
		Requests:               aggregatedProposals.OrderedRequests(requests, *randomness),
		TimeAssumption:         aggregatedProposals.AggregatedTime(),
		Entropy:                *randomness,
		ValidatorFeeTarget:     aggregatedProposals.ValidatorFeeTarget(),
		EstimateGasMode:        false,
		EnableGasBurnLogging:   false,
		MaintenanceModeEnabled: governance.NewStateAccess(chainState).GetMaintenanceStatus(),
		Log:                    c.log.Named("VM"),
	}
	return nil
}

func (c *consImpl) uponVMOutputReceived(vmResult *vm.VMTask) gpa.OutMessages {
	c.output.NeedVMResult = nil
	if len(vmResult.Results) == 0 {
		// No requests were processed, don't have what to do.
		// Will need to retry the consensus with the next log index some time later.
		c.output.Status = Skipped
		c.term.haveOutputProduced()
		return nil
	}

	if vmResult.RotationAddress != nil {
		// Rotation by the Self-Governed Committee.
		essence, err := rotate.MakeRotateStateControllerTransaction(
			vmResult.RotationAddress,
			isc.NewAliasOutputWithID(vmResult.AnchorOutput, vmResult.AnchorOutputID),
			vmResult.TimeAssumption,
			identity.ID{},
			identity.ID{},
		)
		if err != nil {
			c.log.Warnf("cannot create rotation TX, failed to make TX essence: %w", err)
			c.output.Status = Skipped
			c.term.haveOutputProduced()
			return nil
		}
		vmResult.ResultTransactionEssence = essence
		vmResult.StateDraft = nil
	}

	signingMsg, err := vmResult.ResultTransactionEssence.SigningMessage()
	if err != nil {
		panic(fmt.Errorf("uponVMOutputReceived: cannot obtain signing message: %w", err))
	}
	return gpa.NoMessages().
		AddAll(c.subSM.BlockProduced(vmResult.StateDraft)).
		AddAll(c.subTX.VMResultReceived(vmResult)).
		AddAll(c.subDSS.MessageToSignReceived(signingMsg))
}

////////////////////////////////////////////////////////////////////////////////
// TX

// Everything is ready for the output TX, produce it.
func (c *consImpl) uponTXInputsReady(vmResult *vm.VMTask, signature []byte) gpa.OutMessages {
	resultTxEssence := vmResult.ResultTransactionEssence
	resultState := vmResult.StateDraft
	publicKey := c.dkShare.GetSharedPublic()
	var signatureArray [ed25519.SignatureSize]byte
	copy(signatureArray[:], signature)
	signatureForUnlock := &iotago.Ed25519Signature{
		PublicKey: publicKey.AsKey(),
		Signature: signatureArray,
	}
	tx := &iotago.Transaction{
		Essence: resultTxEssence,
		Unlocks: transaction.MakeSignatureAndAliasUnlockFeatures(len(resultTxEssence.Inputs), signatureForUnlock),
	}
	chained, err := transaction.GetAliasOutput(tx, c.chainID.AsAddress())
	if err != nil {
		panic(fmt.Errorf("cannot get AliasOutput from produced TX: %w", err))
	}
	c.output.ResultTransaction = tx
	c.output.ResultNextAliasOutput = chained
	c.output.ResultState = resultState
	c.output.Status = Completed
	c.term.haveOutputProduced()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TERM

func (c *consImpl) uponTerminationCondition() {
	c.output.Terminated = true
}
