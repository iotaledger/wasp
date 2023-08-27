// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// TODO: Cleanup the committees not used for a long time.

// This package implements a protocol for running a chain in a node.
// Its main responsibilities:
//   - Track, which branch is the latest/correct one.
//   - Maintain a set of committee logs (1 for each committee this node participates in).
//   - Maintain a set of consensus instances (one of them is the current one).
//   - Supervise the Mempool and StateMgr.
//   - Handle messages from the NodeConn (AO confirmed / rejected, Request received).
//   - Posting StateTX to NodeConn.
//
// > VARIABLES:
// >     LatestActiveCmt -- The latest committee, that was active.
// >        This field will be nil if the node is not part of the committee.
// >        On the resynchronization it will store the previous active committee.
// >     LatestActiveAO -- The latest AO we are building upon.
// >        Derived, equal to NeedConsensus.BaseAO.
// >     LatestConfirmedAO -- The latest ConfirmedAO from L1.
// >        This one usually follows the LatestAliasOutput,
// >        but can be published from outside and override the LatestAliasOutput.
// >     AccessNodes -- The set of access nodes for the current head.
// >        Union of On-Chain access nodes and the nodes permitted by this node.
// >     NeedConsensus -- A request to run consensus.
// >        Always set based on output of the main CmtLog.
// >     NeedPublishTX -- Requests to publish TX'es.
// >        - Added upon reception of the Consensus Output,
// >          if it is still in NeedConsensus at the time.
// >        - Removed on PublishResult from the NodeConn.
// >
// > UPON Reception of ConfirmedAO:
// >     Set LatestConfirmedAO <- ConfirmedAO
// >     IF this node is in the committee THEN
// >         Pass it to the corresponding CmtLog; HandleCmtLogOutput.
// >     ELSE
// >         IF LatestActiveCmt != nil THEN
// >     	     Send Suspend to Last Active CmtLog; HandleCmtLogOutput
// >         Set LatestActiveCmt <- NIL
// >         Set NeedConsensus <- NIL
// > UPON Reception of PublishResult:
// >     Clear the TX from the NeedPublishTX variable.
// >     If result.confirmed = false THEN
// >         Forward it to ChainMgr; HandleCmtLogOutput.
// >     ELSE
// >         NOP // AO has to be received as ConfirmedAO.
// > UPON Reception of Consensus Output/DONE:
// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// >     Update AccessNodes.
// > UPON Reception of Consensus Output/SKIP:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// > UPON Reception of CmtLog.NextLI message:
// >     Forward it to the corresponding CmtLog; HandleCmtLogOutput.
// >
// > PROCEDURE HandleCmtLogOutput(cmt):
// >     Wrap out messages.
// >     IF cmt == LatestActiveCmt || LatestActiveCmt == NIL THEN
// >         Set LatestActiveCmt <- cmt
// >         Set NeedConsensus <- output.NeedConsensus // Can be nil
// >     ELSE
// >         IF output.NeedConsensus == nil THEN
// >             RETURN // No need to change the committee.
// >         IF LatestActiveCmt != nil THEN
// >             Suspend(LatestActiveCmt)
// >         Set LatestActiveCmt <- cmt
// >         Set NeedConsensus <- output.NeedConsensus
//
// TODO: Why AM is not notified on the committee nodes after rotation?
package chainmanager

import (
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

var ErrNotInCommittee = errors.New("ErrNotInCommittee")

type Output struct {
	cmi *chainMgrImpl
}

func (o *Output) LatestActiveAliasOutput() *isc.AliasOutputWithID {
	if o.cmi.needConsensus == nil {
		return nil
	}
	return o.cmi.needConsensus.BaseAliasOutput
}
func (o *Output) LatestConfirmedAliasOutput() *isc.AliasOutputWithID { return o.cmi.latestConfirmedAO }
func (o *Output) NeedConsensus() *NeedConsensus                      { return o.cmi.needConsensus }
func (o *Output) NeedPublishTX() *shrinkingmap.ShrinkingMap[iotago.TransactionID, *NeedPublishTX] {
	return o.cmi.needPublishTX
}

func (o *Output) String() string {
	return fmt.Sprintf(
		"{chainMgr.Output, LatestConfirmedAliasOutput=%v, NeedConsensus=%v, NeedPublishTX=%v}",
		o.LatestConfirmedAliasOutput(),
		o.NeedConsensus(),
		o.NeedPublishTX(),
	)
}

type NeedConsensus struct {
	CommitteeAddr   iotago.Ed25519Address
	LogIndex        cmt_log.LogIndex
	DKShare         tcrypto.DKShare
	BaseAliasOutput *isc.AliasOutputWithID
}

func (nc *NeedConsensus) IsFor(output *cmt_log.Output) bool {
	return output.GetLogIndex() == nc.LogIndex && output.GetBaseAliasOutput().Equals(nc.BaseAliasOutput)
}

func (nc *NeedConsensus) String() string {
	return fmt.Sprintf(
		"{chainMgr.NeedConsensus, CommitteeAddr=%v, LogIndex=%v, BaseAliasOutput=%v}",
		nc.CommitteeAddr.String(),
		nc.LogIndex,
		nc.BaseAliasOutput,
	)
}

type NeedPublishTX struct {
	CommitteeAddr     iotago.Ed25519Address
	LogIndex          cmt_log.LogIndex
	TxID              iotago.TransactionID
	Tx                *iotago.Transaction
	BaseAliasOutputID iotago.OutputID        // The consumed AliasOutput.
	NextAliasOutput   *isc.AliasOutputWithID // The next one (produced by the TX.)
}

type ChainMgr interface {
	AsGPA() gpa.GPA
}

type cmtLogInst struct {
	committeeAddr iotago.Ed25519Address
	dkShare       tcrypto.DKShare
	gpaInstance   gpa.GPA
	pendingMsgs   []gpa.Message
}

type chainMgrImpl struct {
	chainID                 isc.ChainID                                                      // This instance is responsible for this chain.
	chainStore              state.Store                                                      // Store of the chain state.
	cmtLogs                 map[iotago.Ed25519Address]*cmtLogInst                            // All the committee log instances for this chain.
	consensusStateRegistry  cmt_log.ConsensusStateRegistry                                   // Persistent store for log indexes.
	latestActiveCmt         *iotago.Ed25519Address                                           // The latest active committee.
	latestConfirmedAO       *isc.AliasOutputWithID                                           // The latest confirmed AO (follows Active AO).
	activeNodesCB           func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey)          // All the nodes authorized for being access nodes (for the ActiveAO).
	trackActiveStateCB      func(ao *isc.AliasOutputWithID)                                  // We will call this to set new AO for the active state.
	savePreliminaryBlockCB  func(block state.Block)                                          // We will call this, when a preliminary block matching the tx signatures is received.
	committeeUpdatedCB      func(dkShare tcrypto.DKShare)                                    // Will be called, when a committee changes.
	needConsensus           *NeedConsensus                                                   // Query for a consensus.
	needPublishTX           *shrinkingmap.ShrinkingMap[iotago.TransactionID, *NeedPublishTX] // Query to post TXes.
	dkShareRegistryProvider registry.DKShareRegistryProvider                                 // Source for DKShares.
	varAccessNodeState      VarAccessNodeState
	output                  *Output
	asGPA                   gpa.GPA
	me                      gpa.NodeID
	nodeIDFromPubKey        func(pubKey *cryptolib.PublicKey) gpa.NodeID
	deriveAOByQuorum        bool // Config parameter.
	pipeliningLimit         int  // Config parameter.
	metrics                 *metrics.ChainCmtLogMetrics
	log                     *logger.Logger
}

var (
	_ gpa.GPA  = &chainMgrImpl{}
	_ ChainMgr = &chainMgrImpl{}
)

func New(
	me gpa.NodeID,
	chainID isc.ChainID,
	chainStore state.Store,
	consensusStateRegistry cmt_log.ConsensusStateRegistry,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	activeNodesCB func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey),
	trackActiveStateCB func(ao *isc.AliasOutputWithID),
	savePreliminaryBlockCB func(block state.Block),
	committeeUpdatedCB func(dkShare tcrypto.DKShare),
	deriveAOByQuorum bool,
	pipeliningLimit int,
	metrics *metrics.ChainCmtLogMetrics,
	log *logger.Logger,
) (ChainMgr, error) {
	cmi := &chainMgrImpl{
		chainID:                 chainID,
		chainStore:              chainStore,
		cmtLogs:                 map[iotago.Ed25519Address]*cmtLogInst{},
		consensusStateRegistry:  consensusStateRegistry,
		activeNodesCB:           activeNodesCB,
		trackActiveStateCB:      trackActiveStateCB,
		savePreliminaryBlockCB:  savePreliminaryBlockCB,
		committeeUpdatedCB:      committeeUpdatedCB,
		needConsensus:           nil,
		needPublishTX:           shrinkingmap.New[iotago.TransactionID, *NeedPublishTX](),
		dkShareRegistryProvider: dkShareRegistryProvider,
		varAccessNodeState:      NewVarAccessNodeState(chainID, log.Named("VAS")),
		me:                      me,
		nodeIDFromPubKey:        nodeIDFromPubKey,
		deriveAOByQuorum:        deriveAOByQuorum,
		pipeliningLimit:         pipeliningLimit,
		metrics:                 metrics,
		log:                     log,
	}
	cmi.output = &Output{cmi: cmi}
	cmi.asGPA = gpa.NewOwnHandler(me, cmi)
	return cmi, nil
}

// Implements the CmtLog interface.
func (cmi *chainMgrImpl) AsGPA() gpa.GPA {
	return cmi.asGPA
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Input(input gpa.Input) gpa.OutMessages {
	switch input := input.(type) {
	case *inputAliasOutputConfirmed:
		return cmi.handleInputAliasOutputConfirmed(input)
	case *inputChainTxPublishResult:
		return cmi.handleInputChainTxPublishResult(input)
	case *inputConsensusOutputDone:
		return cmi.handleInputConsensusOutputDone(input)
	case *inputConsensusOutputSkip:
		return cmi.handleInputConsensusOutputSkip(input)
	case *inputConsensusTimeout:
		return cmi.handleInputConsensusTimeout(input)
	case *inputCanPropose:
		return cmi.handleInputCanPropose()
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msg := msg.(type) {
	case *msgCmtLog:
		return cmi.handleMsgCmtLog(msg)
	case *msgBlockProduced:
		return cmi.handleMsgBlockProduced(msg)
	}
	panic(fmt.Errorf("unexpected message %T: %+v", msg, msg))
}

// > UPON Reception of ConfirmedAO:
// >     Set LatestConfirmedAO <- ConfirmedAO
// >     IF this node is in the committee THEN
// >         Pass it to the corresponding CmtLog; HandleCmtLogOutput(ConfirmedAO.Cmt).
// >     ELSE
// >         IF LatestActiveCmt != nil THEN
// >     	     Send Suspend to Last Active CmtLog; HandleCmtLogOutput(LatestActiveCmt)
// >         Set LatestActiveCmt <- NIL
// >         Set NeedConsensus <- NIL
func (cmi *chainMgrImpl) handleInputAliasOutputConfirmed(input *inputAliasOutputConfirmed) gpa.OutMessages {
	cmi.log.Debugf("handleInputAliasOutputConfirmed: %+v", input)
	//
	// >     Set LatestConfirmedAO <- ConfirmedAO
	vsaTip, vsaUpdated := cmi.varAccessNodeState.BlockConfirmed(input.aliasOutput)
	cmi.latestConfirmedAO = input.aliasOutput
	msgs := gpa.NoMessages()
	committeeAddr := input.aliasOutput.GetAliasOutput().StateController().(*iotago.Ed25519Address)
	committeeLog, err := cmi.ensureCmtLog(*committeeAddr)
	if errors.Is(err, ErrNotInCommittee) {
		// >     IF this node is in the committee THEN ... ELSE
		// >         IF LatestActiveCmt != nil THEN
		// >     	     Send Suspend to Last Active CmtLog; HandleCmtLogOutput(LatestActiveCmt)
		// >         Set LatestActiveCmt <- NIL
		// >         Set NeedConsensus <- NIL
		if cmi.latestActiveCmt != nil {
			msgs.AddAll(cmi.suspendCommittee(cmi.latestActiveCmt))
			cmi.committeeUpdatedCB(nil)
			cmi.latestActiveCmt = nil
		}
		cmi.needConsensus = nil
		if vsaUpdated && vsaTip != nil {
			cmi.log.Debugf("⊢ going to track %v as an access node on confirmed block.", vsaTip)
			cmi.trackActiveStateCB(vsaTip)
		}
		cmi.log.Debugf("This node is not in the committee for aliasOutput: %v", input.aliasOutput)
		return msgs
	}
	if err != nil {
		cmi.log.Warnf("Failed to get CmtLog: %v", err)
		return msgs
	}
	// >     IF this node is in the committee THEN
	// >         Pass it to the corresponding CmtLog; HandleCmtLogOutput.
	msgs.AddAll(cmi.handleCmtLogOutput(
		committeeLog,
		committeeLog.gpaInstance.Input(cmt_log.NewInputAliasOutputConfirmed(input.aliasOutput)),
	))
	return msgs
}

// > UPON Reception of PublishResult:
// >     Clear the TX from the NeedPublishTX variable.
// >     If result.confirmed = false THEN
// >         Forward it to ChainMgr; HandleCmtLogOutput.
// >     ELSE
// >         NOP // AO has to be received as Confirmed AO.
func (cmi *chainMgrImpl) handleInputChainTxPublishResult(input *inputChainTxPublishResult) gpa.OutMessages {
	cmi.log.Debugf("handleInputChainTxPublishResult: %+v", input)
	// >     Clear the TX from the NeedPublishTX variable.
	cmi.needPublishTX.Delete(input.txID)
	if input.confirmed {
		// >     If result.confirmed = false THEN ... ELSE
		// >         NOP // AO has to be received as Confirmed AO. // TODO: Not true, anymore.
		return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
			return cl.Input(cmt_log.NewInputConsensusOutputConfirmed(input.aliasOutput, input.logIndex))
		})
	}
	// >     If result.confirmed = false THEN
	// >         Forward it to ChainMgr; HandleCmtLogOutput.
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmt_log.NewInputConsensusOutputRejected(input.aliasOutput, input.logIndex))
	})
}

// > UPON Reception of Consensus Output/DONE:
// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// >     Update AccessNodes.
func (cmi *chainMgrImpl) handleInputConsensusOutputDone(input *inputConsensusOutputDone) gpa.OutMessages {
	cmi.log.Debugf("handleInputConsensusOutputDone: %+v", input)
	msgs := gpa.NoMessages()
	// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
	// >         Add ConsensusOutput.TX to NeedPublishTX
	if true { // TODO: Reconsider this condition. Several recent consensus instances should be published, if we run consensus instances in parallel.
		txID := input.consensusResult.NextAliasOutput.TransactionID()
		if !cmi.needPublishTX.Has(txID) && input.consensusResult.Block != nil {
			// Inform the access nodes on new block produced.
			block := input.consensusResult.Block
			activeAccessNodes, activeCommitteeNodes := cmi.activeNodesCB()
			cmi.log.Debugf(
				"Sending MsgBlockProduced (stateIndex=%v, l1Commitment=%v, txID=%v) to access nodes: %v except committeeNodes %v",
				block.StateIndex(), block.L1Commitment(), txID.ToHex(), activeAccessNodes, activeCommitteeNodes,
			)
			for i := range activeAccessNodes {
				if lo.Contains(activeCommitteeNodes, activeAccessNodes[i]) {
					continue
				}
				msgs.Add(NewMsgBlockProduced(cmi.nodeIDFromPubKey(activeAccessNodes[i]), input.consensusResult.Transaction, block))
			}
		}
		cmi.needPublishTX.Set(txID, &NeedPublishTX{
			CommitteeAddr:     input.committeeAddr,
			LogIndex:          input.logIndex,
			TxID:              txID,
			Tx:                input.consensusResult.Transaction,
			BaseAliasOutputID: input.consensusResult.BaseAliasOutput,
			NextAliasOutput:   input.consensusResult.NextAliasOutput,
		})
	}
	//
	// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
	msgs.AddAll(cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmt_log.NewInputConsensusOutputDone(input.logIndex, input.proposedBaseAO, input.consensusResult.BaseAliasOutput, input.consensusResult.NextAliasOutput))
	}))
	return msgs
}

// > UPON Reception of Consensus Output/SKIP:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
func (cmi *chainMgrImpl) handleInputConsensusOutputSkip(input *inputConsensusOutputSkip) gpa.OutMessages {
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmt_log.NewInputConsensusOutputSkip(input.logIndex, input.proposedBaseAO))
	})
}

// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
func (cmi *chainMgrImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	cmi.log.Debugf("handleInputConsensusTimeout: %+v", input)
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmt_log.NewInputConsensusTimeout(input.logIndex))
	})
}

func (cmi *chainMgrImpl) handleInputCanPropose() gpa.OutMessages {
	cmi.log.Debugf("handleInputCanPropose")
	return cmi.withAllCmtLogs(func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmt_log.NewInputCanPropose())
	})
}

// > UPON Reception of CmtLog.NextLI message:
// >     Forward it to the corresponding CmtLog; HandleCmtLogOutput.
func (cmi *chainMgrImpl) handleMsgCmtLog(msg *msgCmtLog) gpa.OutMessages {
	cmi.log.Debugf("handleMsgCmtLog: %+v", msg)
	return cmi.withCmtLog(msg.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Message(msg.wrapped)
	})
}

func (cmi *chainMgrImpl) handleMsgBlockProduced(msg *msgBlockProduced) gpa.OutMessages {
	cmi.log.Debugf("handleMsgBlockProduced: %+v", msg)
	vsaTip, vsaUpdated, l1Commitment := cmi.varAccessNodeState.BlockProduced(msg.tx)
	//
	// Save the block, if it matches all the signatures by the current committee.
	// This will save us a round-trip to query the block from the sender.
	if l1Commitment != nil {
		if msg.block.L1Commitment().Equals(l1Commitment) {
			cmi.savePreliminaryBlockCB(msg.block)
		} else {
			cmi.log.Warnf("Received msgBlockProduced, but publishedAO.l1Commitment != block.l1Commitment.")
		}
	}
	//
	// Update the active state, if needed.
	if vsaUpdated && vsaTip != nil && cmi.latestActiveCmt == nil {
		cmi.log.Debugf("⊢ going to track %v as an access node on unconfirmed block.", vsaTip)
		cmi.trackActiveStateCB(vsaTip)
	}
	return nil
}

// > PROCEDURE HandleCmtLogOutput(cmt):
// >     Wrap out messages.
// >     IF cmt == LatestActiveCmt || LatestActiveCmt == NIL THEN
// >         Set LatestActiveCmt <- cmt
// >         Set NeedConsensus <- output.NeedConsensus // Can be nil
// >     ELSE
// >         IF output.NeedConsensus == nil THEN
// >             RETURN // No need to change the committee.
// >         IF LatestActiveCmt != nil THEN
// >             Suspend(LatestActiveCmt)
// >         Set LatestActiveCmt <- cmt
// >         Set NeedConsensus <- output.NeedConsensus
func (cmi *chainMgrImpl) handleCmtLogOutput(cli *cmtLogInst, cliMsgs gpa.OutMessages) gpa.OutMessages {
	//
	// >     Wrap out messages.
	msgs := gpa.NoMessages()
	msgs.AddAll(cmi.wrapCmtLogMsgs(cli, cliMsgs))
	outputUntyped := cli.gpaInstance.Output()
	// >     IF cmt == LatestActiveCmt || LatestActiveCmt == NIL THEN
	// >         Set LatestActiveCmt <- cmt
	// >         Set NeedConsensus <- output.NeedConsensus // Can be nil
	if cmi.latestActiveCmt == nil || cli.committeeAddr.Equal(cmi.latestActiveCmt) {
		cmi.committeeUpdatedCB(cli.dkShare)
		cmi.ensureNeedConsensus(cli, outputUntyped)
		cmi.latestActiveCmt = &cli.committeeAddr
		return msgs
	}
	// >     ELSE
	// >         IF output.NeedConsensus == nil THEN
	// >             RETURN // No need to change the committee.
	// >         IF LatestActiveCmt != nil THEN
	// >             Suspend(LatestActiveCmt)
	// >         Set LatestActiveCmt <- cmt
	// >         Set NeedConsensus <- output.NeedConsensus
	if outputUntyped == nil {
		return msgs
	}
	if !cmi.latestActiveCmt.Equal(&cli.committeeAddr) {
		msgs.AddAll(cmi.suspendCommittee(cmi.latestActiveCmt))
		cmi.committeeUpdatedCB(cli.dkShare)
		cmi.latestActiveCmt = &cli.committeeAddr
	}
	cmi.ensureNeedConsensus(cli, outputUntyped)
	return msgs
}

func (cmi *chainMgrImpl) ensureNeedConsensus(cli *cmtLogInst, outputUntyped gpa.Output) {
	if outputUntyped == nil {
		cmi.needConsensus = nil
		return
	}
	output := outputUntyped.(*cmt_log.Output)
	if cmi.needConsensus != nil && cmi.needConsensus.IsFor(output) {
		// Not changed, keep it.
		return
	}
	committeeAddress := output.GetBaseAliasOutput().GetStateAddress()
	dkShare, err := cmi.dkShareRegistryProvider.LoadDKShare(committeeAddress)
	if errors.Is(err, tcrypto.ErrDKShareNotFound) {
		// Rotated to other nodes, so we don't need to start the next consensus.
		cmi.needConsensus = nil
		return
	}
	if err != nil {
		panic(fmt.Errorf("ensureNeedConsensus cannot load DKShare for %v: %w", committeeAddress, err))
	}
	cmi.needConsensus = &NeedConsensus{
		CommitteeAddr:   cli.committeeAddr,
		LogIndex:        output.GetLogIndex(),
		DKShare:         dkShare,
		BaseAliasOutput: output.GetBaseAliasOutput(),
	}
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Output() gpa.Output {
	return cmi.output
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) StatusString() string { // TODO: Call it periodically. Show the active committee.
	return fmt.Sprintf("{ChainMgr,confirmedAO=%v,activeAO=%v}",
		cmi.output.LatestConfirmedAliasOutput().String(),
		cmi.output.LatestActiveAliasOutput().String(),
	)
}

////////////////////////////////////////////////////////////////////////////////
// Helper functions.

func (cmi *chainMgrImpl) wrapCmtLogMsgs(cli *cmtLogInst, outMsgs gpa.OutMessages) gpa.OutMessages {
	wrappedMsgs := gpa.NoMessages()
	outMsgs.MustIterate(func(msg gpa.Message) {
		wrappedMsgs.Add(NewMsgCmtLog(cli.committeeAddr, msg))
	})
	return wrappedMsgs
}

func (cmi *chainMgrImpl) suspendCommittee(committeeAddr *iotago.Ed25519Address) gpa.OutMessages {
	for ca, cli := range cmi.cmtLogs {
		if !ca.Equal(committeeAddr) {
			continue
		}
		return cmi.wrapCmtLogMsgs(cli, cli.gpaInstance.Input(cmt_log.NewInputSuspend()))
	}
	return nil
}

func (cmi *chainMgrImpl) withCmtLog(committeeAddr iotago.Ed25519Address, handler func(cl gpa.GPA) gpa.OutMessages) gpa.OutMessages {
	cli, err := cmi.ensureCmtLog(committeeAddr)
	if err != nil {
		cmi.log.Warnf("cannot find committee: %v", committeeAddr)
		return nil
	}
	return gpa.NoMessages().AddAll(cmi.handleCmtLogOutput(cli, handler(cli.gpaInstance)))
}

func (cmi *chainMgrImpl) withAllCmtLogs(handler func(cl gpa.GPA) gpa.OutMessages) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, cli := range cmi.cmtLogs {
		msgs.AddAll(cmi.handleCmtLogOutput(cli, handler(cli.gpaInstance)))
	}
	return msgs
}

// NOTE: ErrNotInCommittee
func (cmi *chainMgrImpl) ensureCmtLog(committeeAddr iotago.Ed25519Address) (*cmtLogInst, error) {
	if cli, ok := cmi.cmtLogs[committeeAddr]; ok {
		return cli, nil
	}
	//
	// Create a committee if not created yet.
	dkShare, err := cmi.dkShareRegistryProvider.LoadDKShare(&committeeAddr)
	if errors.Is(err, tcrypto.ErrDKShareNotFound) {
		return nil, ErrNotInCommittee
	}
	if err != nil {
		return nil, fmt.Errorf("ensureCmtLog cannot load DKShare for committeeAddress=%v: %w", committeeAddr, err)
	}

	clInst, err := cmt_log.New(
		cmi.me, cmi.chainID, dkShare, cmi.consensusStateRegistry, cmi.nodeIDFromPubKey, cmi.deriveAOByQuorum, cmi.pipeliningLimit, cmi.metrics,
		cmi.log.Named(fmt.Sprintf("CL-%v", dkShare.GetSharedPublic().AsEd25519Address().String()[:10])),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create cmtLog for committeeAddress=%v: %w", committeeAddr, err)
	}
	clGPA := clInst.AsGPA()
	cli := &cmtLogInst{
		committeeAddr: committeeAddr,
		dkShare:       dkShare,
		gpaInstance:   clGPA,
		pendingMsgs:   []gpa.Message{},
	}
	cmi.cmtLogs[committeeAddr] = cli
	return cli, nil
}
