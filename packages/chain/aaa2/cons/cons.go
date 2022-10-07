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
// >         Submit the result hash to the DSS.
// >     ELSE
// >         OUTPUT SKIP
// > UPON Reception of VM Result and a signature from the DSS
// >     IF rotation THEN
// >        OUTPUT Signed Governance TX.
// >     ELSE
// >        OUTPUT Signed State Transition TX
//
// We move all the synchronization logic to separate objects (upon_...). They are
// responsible for waiting specific data and then triggering the next state action
// once. This way we hope to solve a lot of race conditions gracefully. The `upon`
// predicates and the corresponding done functions should not depend on each other.
// If some data is needed at several places, it should be passed to several predicates.
package cons

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/identity"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/bp"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemACS"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemDSS"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemMP"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemRND"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemSM"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemTX"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/subsystemVM"
	"github.com/iotaledger/wasp/packages/chain/dss"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acs"
	"github.com/iotaledger/wasp/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/packages/gpa/cc/semi"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
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

type OutputState byte

const (
	Running   OutputState = iota // Instance is still running.
	Completed                    // Consensus reached, TX is prepared for publication.
	Skipped                      // Consensus reached, no TX should be posted for this LogIndex.
)

type Output struct {
	State      OutputState
	Terminated bool
	//
	// Requests for other components.
	NeedMempoolProposal       *isc.AliasOutputWithID // Requests for the mempool are needed for this Base Alias Output.
	NeedMempoolRequests       []*isc.RequestRef      // Request payloads are needed from mempool for this IDs/Hash.
	NeedStateMgrStateProposal *isc.AliasOutputWithID // Query for a proposal for Virtual State (it will go to the batch proposal).
	NeedStateMgrDecidedState  *iotago.OutputID       // Query for a decided Virtual State to be used by VM.
	NeedVMResult              *vm.VMTask             // VM Result is needed for this (agreed) batch.
	//
	// Following is the final result.
	// All the fields are filled, if State == Completed.
	ResultTransaction     *iotago.Transaction
	ResultNextAliasOutput *isc.AliasOutputWithID
	ResultState           state.VirtualStateAccess
}

type consImpl struct {
	chainID        isc.ChainID
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
	subMP          *subsystemMP.SubsystemMP   // Mempool.
	subSM          *subsystemSM.SubsystemSM   // StateMgr.
	subDSS         *subsystemDSS.SubsystemDSS // Distributed Schnorr Signature.
	subACS         *subsystemACS.SubsystemACS // Asynchronous Common Subset.
	subRND         *subsystemRND.SubsystemRND // Randomness.
	subVM          *subsystemVM.SubsystemVM   // Virtual Machine.
	subTX          *subsystemTX.SubsystemTX   // Building final TX.
	term           *termCondition             // To detect, when this instance can be terminated.
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
			panic(xerrors.Errorf("cannot convert nodePK[%v] to kyber.Point: %w", i, err))
		}
	}

	f := len(dkShareNodePubKeys) - int(dkShare.GetT())
	myKyberKeys, err := mySK.AsKyberKeyPair()
	if err != nil {
		panic(xerrors.Errorf("cannot convert node's SK to kyber.Scalar: %w", err))
	}
	longTermDKS := dkShare.DSSSecretShare()
	acsCCInstFunc := func(nodeID gpa.NodeID, round int) gpa.GPA {
		var roundBin [4]byte
		binary.BigEndian.PutUint32(roundBin[:], uint32(round))
		sid := hashing.HashDataBlake2b(instID, []byte(nodeID), roundBin[:]).Bytes()
		realCC := blssig.New(blsSuite, nodeIDs, dkShare.BLSCommits(), dkShare.BLSPriShare(), int(dkShare.BLSThreshold()), me, sid, log)
		return semi.New(round, realCC)
	}
	c := &consImpl{
		chainID:        chainID,
		edSuite:        edSuite,
		blsSuite:       blsSuite,
		dkShare:        dkShare,
		processorCache: processorCache,
		nodeIDs:        nodeIDs,
		me:             me,
		f:              f,
		dss:            dss.New(edSuite, nodeIDs, nodePKs, f, me, myKyberKeys.Private, longTermDKS, log),
		acs:            acs.New(nodeIDs, me, f, acsCCInstFunc, log),
		output:         &Output{State: Running},
		log:            log,
	}
	c.asGPA = gpa.NewOwnHandler(me, c)
	c.msgWrapper = gpa.NewMsgWrapper(msgTypeWrapped, c.msgWrapperFunc)
	c.subMP = subsystemMP.New(
		c.uponMPProposalInputsReady,
		c.uponMPProposalReceived,
		c.uponMPRequestsNeeded,
		c.uponMPRequestsReceived,
	)
	c.subSM = subsystemSM.New(
		c.uponSMStateProposalQueryInputsReady,
		c.uponSMStateProposalReceived,
		c.uponSMDecidedStateQueryInputsReady,
		c.uponSMDecidedStateReceived,
	)
	c.subDSS = subsystemDSS.New(
		c.uponDSSInitialInputsReady,
		c.uponDSSIndexProposalReady,
		c.uponDSSSigningInputsReceived,
		c.uponDSSOutputReady,
	)
	c.subACS = subsystemACS.New(
		c.uponACSInputsReceived,
		c.uponACSOutputReceived,
		c.uponACSTerminated,
	)
	c.subRND = subsystemRND.New(
		int(dkShare.BLSThreshold()),
		c.uponRNDInputsReady,
		c.uponRNDSigSharesReady,
	)
	c.subVM = subsystemVM.New(
		c.uponVMInputsReceived,
		c.uponVMOutputReceived,
	)
	c.subTX = subsystemTX.New(
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
			return nil, xerrors.Errorf("unexpected DSS index: %v", index)
		}
		return c.dss.AsGPA(), nil
	}
	if subsystem == subsystemTypeACS {
		if index != 0 {
			return nil, xerrors.Errorf("unexpected ACS index: %v", index)
		}
		return c.acs.AsGPA(), nil
	}
	return nil, xerrors.Errorf("unexpected subsystem: %v", subsystem)
}

func (c *consImpl) AsGPA() gpa.GPA {
	return c.asGPA
}

func (c *consImpl) Input(input gpa.Input) gpa.OutMessages {
	if baseAliasOutput, ok := input.(*isc.AliasOutputWithID); ok {
		msgs := gpa.NoMessages()
		msgs.AddAll(c.subMP.BaseAliasOutputReceived(baseAliasOutput))
		msgs.AddAll(c.subSM.ProposedBaseAliasOutputReceived(baseAliasOutput))
		msgs.AddAll(c.subDSS.InitialInputReceived())
		return msgs
	}
	panic(xerrors.Errorf("unexpected input: %v", input))
}

// Implements the gpa.GPA interface.
// Here we route all the messages.
func (c *consImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msgT := msg.(type) {
	case *msgMempoolProposal:
		return c.subMP.ProposalReceived(msgT.requestRefs)
	case *msgMempoolRequests:
		return c.subMP.RequestsReceived(msgT.requests)
	case *msgStateMgrProposalConfirmed:
		return c.subSM.StateProposalConfirmedByStateMgr(msgT.baseAliasOutput)
	case *msgStateMgrDecidedVirtualState:
		return c.subSM.DecidedVirtualStateReceived(msgT.aliasOutput, msgT.stateBaseline, msgT.virtualStateAccess)
	case *msgTimeData:
		return c.subACS.TimeDataReceived(msgT.timeData)
	case *msgBLSPartialSig:
		return c.subRND.BLSPartialSigReceived(msgT.sender, msgT.partialSig)
	case *msgVMResult:
		return c.subVM.VMResultReceived(msgT.task)
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
		panic(xerrors.Errorf("unexpected subsystem after check: %+v", msg))
	}
	panic(xerrors.Errorf("unexpected message: %v", msg))
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

func (c *consImpl) uponSMDecidedStateQueryInputsReady(decidedBaseAliasOutputID *iotago.OutputID) gpa.OutMessages {
	c.output.NeedStateMgrDecidedState = decidedBaseAliasOutputID
	return nil
}

func (c *consImpl) uponSMDecidedStateReceived(aliasOutput *isc.AliasOutputWithID, stateBaseline coreutil.StateBaseline, virtualStateAccess state.VirtualStateAccess) gpa.OutMessages {
	c.output.NeedStateMgrDecidedState = nil
	return c.subVM.DecidedStateReceived(aliasOutput, stateBaseline, virtualStateAccess)
}

////////////////////////////////////////////////////////////////////////////////
// DSS

func (c *consImpl) uponDSSInitialInputsReady() gpa.OutMessages {
	sub, subMsgs, err := c.msgWrapper.DelegateInput(subsystemTypeDSS, 0, nil)
	if err != nil {
		panic(xerrors.Errorf("cannot provide input to DSS: %w", err))
	}
	return gpa.NoMessages().
		AddAll(subMsgs).
		AddAll(c.subDSS.DSSOutputReceived(sub.Output()))
}

func (c *consImpl) uponDSSIndexProposalReady(indexProposal []int) gpa.OutMessages {
	return c.subACS.DSSIndexProposalReceived(indexProposal)
}

func (c *consImpl) uponDSSSigningInputsReceived(sub *subsystemDSS.SubsystemDSS) gpa.OutMessages {
	dssMsg := c.dss.NewMsgDecided(sub.DecidedIndexProposals, sub.MessageToSign)
	subDSS, subMsgs, err := c.msgWrapper.DelegateMessage(gpa.NewWrappingMsg(msgTypeWrapped, subsystemTypeDSS, 0, dssMsg))
	if err != nil {
		panic(xerrors.Errorf("cannot provide inputs for signing: %w", err))
	}
	return gpa.NoMessages().AddAll(subMsgs).AddAll(c.subDSS.DSSOutputReceived(subDSS.Output()))
}

func (c *consImpl) uponDSSOutputReady(signature []byte) gpa.OutMessages {
	return c.subTX.SignatureReceived(signature)
}

////////////////////////////////////////////////////////////////////////////////
// ACS

func (c *consImpl) uponACSInputsReceived(sub *subsystemACS.SubsystemACS) gpa.OutMessages {
	batchProposal := bp.NewBatchProposal(
		*c.dkShare.GetIndex(),
		sub.BaseAliasOutput.ID(),
		util.NewFixedSizeBitVector(int(c.dkShare.GetN())).SetBits(sub.DSSIndexProposal),
		sub.TimeData,
		isc.NewContractAgentID(&c.chainID, 0),
		sub.RequestRefs,
	)
	subACS, subMsgs, err := c.msgWrapper.DelegateInput(subsystemTypeACS, 0, batchProposal.Bytes())
	if err != nil {
		panic(xerrors.Errorf("cannot provide input to the ACS: %w", err))
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
		c.output.State = Skipped
		c.term.haveOutputProduced()
		return nil
	}
	baoID := *aggr.DecidedBaseAliasOutputID()
	return gpa.NoMessages().
		AddAll(c.subMP.RequestsNeeded(aggr.DecidedRequestRefs())).
		AddAll(c.subSM.DecidedVirtualStateNeeded(&baoID)).
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
		panic(xerrors.Errorf("cannot sign share for randomness: %w", err))
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

func (c *consImpl) uponVMInputsReceived(sub *subsystemVM.SubsystemVM) gpa.OutMessages {
	// The decided base alias output can be different from that we have proposed!
	decidedBaseAliasOutputID := sub.AggregatedProposals.DecidedBaseAliasOutputID()
	c.output.NeedVMResult = &vm.VMTask{
		ACSSessionID:           0, // TODO: Remove the ACSSessionID when old consensus Impl is removed.
		Processors:             c.processorCache,
		AnchorOutput:           sub.BaseAliasOutput.GetAliasOutput(),
		AnchorOutputID:         *decidedBaseAliasOutputID,
		SolidStateBaseline:     sub.StateBaseline,
		Requests:               sub.Requests,
		TimeAssumption:         sub.AggregatedProposals.AggregatedTime(),
		Entropy:                *sub.Randomness,
		ValidatorFeeTarget:     sub.AggregatedProposals.ValidatorFeeTarget(),
		EstimateGasMode:        false,
		EnableGasBurnLogging:   false,
		VirtualStateAccess:     sub.VirtualStateAccess.Copy(),
		MaintenanceModeEnabled: governance.NewStateAccess(sub.VirtualStateAccess.KVStore()).GetMaintenanceStatus(),
		Log:                    c.log.Named("VM"),
	}
	return nil
}

func (c *consImpl) uponVMOutputReceived(vmResult *vm.VMTask) gpa.OutMessages {
	c.output.NeedVMResult = nil
	if len(vmResult.Results) == 0 {
		// No requests were processed, don't have what to do.
		// Will need to retry the consensus with the next log index some time later.
		c.output.State = Skipped
		c.term.haveOutputProduced()
		return nil
	}
	signingMsg, err := vmResult.ResultTransactionEssence.SigningMessage()
	if err != nil {
		panic(xerrors.Errorf("uponVMOutputReceived: cannot obtain signing message: %v", err))
	}
	msgs := gpa.NoMessages()
	msgs.AddAll(c.subTX.VMResultReceived(vmResult))
	msgs.AddAll(c.subDSS.MessageToSignReceived(signingMsg))
	return msgs
}

////////////////////////////////////////////////////////////////////////////////
// TX

// Everything is ready for the output TX, produce it.
func (c *consImpl) uponTXInputsReady(sub *subsystemTX.SubsystemTX) gpa.OutMessages {
	vmResult := sub.VMResult
	var resultTxEssence *iotago.TransactionEssence
	var resultState state.VirtualStateAccess
	if vmResult.RotationAddress != nil {
		// Rotation by the Self-Governed Committee.
		essence, err := rotate.MakeRotateStateControllerTransaction(
			vmResult.RotationAddress,
			isc.NewAliasOutputWithID(vmResult.AnchorOutput, vmResult.AnchorOutputID.UTXOInput()),
			vmResult.TimeAssumption,
			identity.ID{},
			identity.ID{},
		)
		if err != nil {
			c.log.Warnf("cannot create rotation TX, failed to make TX essence: %v", err)
			c.output.State = Skipped
			c.term.haveOutputProduced()
			return nil
		}
		resultTxEssence = essence
		resultState = nil
	} else {
		if vmResult.ResultTransactionEssence == nil {
			c.log.Warnf("cannot create state TX, failed to get TX essence: nil")
		}
		resultTxEssence = vmResult.ResultTransactionEssence
		resultState = vmResult.VirtualStateAccess
	}
	publicKey := c.dkShare.GetSharedPublic()
	var signatureArray [ed25519.SignatureSize]byte
	copy(signatureArray[:], sub.Signature)
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
		panic(xerrors.Errorf("cannot get AliasOutput from produced TX: %w", err))
	}
	c.output.ResultTransaction = tx
	c.output.ResultNextAliasOutput = chained
	c.output.ResultState = resultState
	c.output.State = Completed
	c.term.haveOutputProduced()
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TERM

func (c *consImpl) uponTerminationCondition() {
	c.output.Terminated = true
}
