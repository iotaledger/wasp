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
package chainMgr

import (
	"errors"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
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
func (o *Output) LatestConfirmedAliasOutput() *isc.AliasOutputWithID     { return o.cmi.latestConfirmedAO }
func (o *Output) ActiveAccessNodes() []*cryptolib.PublicKey              { return o.cmi.activeAccessNodes }
func (o *Output) NeedConsensus() *NeedConsensus                          { return o.cmi.needConsensus }
func (o *Output) NeedPublishTX() map[iotago.TransactionID]*NeedPublishTX { return o.cmi.needPublishTX }

type NeedConsensus struct {
	CommitteeAddr   iotago.Ed25519Address
	LogIndex        cmtLog.LogIndex
	DKShare         tcrypto.DKShare
	BaseAliasOutput *isc.AliasOutputWithID
}

func (nc *NeedConsensus) IsFor(output *cmtLog.Output) bool {
	return output.GetLogIndex() == nc.LogIndex && output.GetBaseAliasOutput().Equals(nc.BaseAliasOutput)
}

type NeedPublishTX struct {
	CommitteeAddr     iotago.Ed25519Address
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
	gpaInstance   gpa.GPA
	pendingMsgs   []gpa.Message
}

type chainMgrImpl struct {
	chainID                 isc.ChainID                             // This instance is responsible for this chain.
	cmtLogs                 map[iotago.Ed25519Address]*cmtLogInst   // All the committee log instances for this chain.
	cmtLogStore             cmtLog.Store                            // Persistent store for log indexes.
	latestActiveCmt         *iotago.Ed25519Address                  // The latest active committee.
	latestConfirmedAO       *isc.AliasOutputWithID                  // The latest confirmed AO (follows Active AO).
	activeAccessNodes       []*cryptolib.PublicKey                  // All the nodes authorized for being access nodes (for the ActiveAO).
	activeAccessNodesCB     func([]*cryptolib.PublicKey)            // Called, when a list of access nodes has changed.
	needConsensus           *NeedConsensus                          // Query for a consensus.
	needPublishTX           map[iotago.TransactionID]*NeedPublishTX // Query to post TXes.
	dkShareRegistryProvider registry.DKShareRegistryProvider        // Source for DKShares.
	output                  *Output
	asGPA                   gpa.GPA
	me                      gpa.NodeID
	nodeIDFromPubKey        func(pubKey *cryptolib.PublicKey) gpa.NodeID
	log                     *logger.Logger
}

var (
	_ gpa.GPA  = &chainMgrImpl{}
	_ ChainMgr = &chainMgrImpl{}
)

func New(
	me gpa.NodeID,
	chainID isc.ChainID,
	cmtLogStore cmtLog.Store,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	activeAccessNodesCB func([]*cryptolib.PublicKey),
	log *logger.Logger,
) (ChainMgr, error) {
	cmi := &chainMgrImpl{
		chainID:                 chainID,
		cmtLogs:                 map[iotago.Ed25519Address]*cmtLogInst{},
		cmtLogStore:             cmtLogStore,
		activeAccessNodes:       []*cryptolib.PublicKey{},
		activeAccessNodesCB:     activeAccessNodesCB,
		needConsensus:           nil,
		needPublishTX:           map[iotago.TransactionID]*NeedPublishTX{},
		dkShareRegistryProvider: dkShareRegistryProvider,
		me:                      me,
		nodeIDFromPubKey:        nodeIDFromPubKey,
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
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Message(msg gpa.Message) gpa.OutMessages {
	msgCL, ok := msg.(*msgCmtLog)
	if !ok {
		panic(xerrors.Errorf("unexpected message %T: %+v", msg, msg))
	}
	return cmi.handleMsgCmtLog(msgCL)
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
	cmi.latestConfirmedAO = input.aliasOutput
	msgs := gpa.NoMessages()
	committeeAddr := input.aliasOutput.GetAliasOutput().StateController().(*iotago.Ed25519Address)
	committeeLog, cmtMsgs, err := cmi.ensureCmtLog(*committeeAddr)
	msgs.AddAll(cmtMsgs)
	if errors.Is(err, ErrNotInCommittee) {
		// >     IF this node is in the committee THEN ... ELSE
		// >         IF LatestActiveCmt != nil THEN
		// >     	     Send Suspend to Last Active CmtLog; HandleCmtLogOutput(LatestActiveCmt)
		// >         Set LatestActiveCmt <- NIL
		// >         Set NeedConsensus <- NIL
		if cmi.latestActiveCmt != nil {
			msgs.AddAll(cmi.suspendCommittee(cmi.latestActiveCmt))
		}
		cmi.latestActiveCmt = nil
		cmi.needConsensus = nil
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
		committeeLog.gpaInstance.Input(cmtLog.NewInputAliasOutputConfirmed(input.aliasOutput)),
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
	delete(cmi.needPublishTX, input.txID)
	if input.confirmed {
		// >     If result.confirmed = false THEN ... ELSE
		// >         NOP // AO has to be received as Confirmed AO.
		return nil
	}
	// >     If result.confirmed = false THEN
	// >         Forward it to ChainMgr; HandleCmtLogOutput.
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmtLog.NewInputAliasOutputRejected(input.aliasOutput))
	})
}

// > UPON Reception of Consensus Output/DONE:
// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// >     Update AccessNodes.
func (cmi *chainMgrImpl) handleInputConsensusOutputDone(input *inputConsensusOutputDone) gpa.OutMessages {
	cmi.log.Debugf("handleInputConsensusOutputDone: %+v", input)
	// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
	// >         Add ConsensusOutput.TX to NeedPublishTX
	if cmi.needConsensus.BaseAliasOutput.OutputID() == input.baseAliasOutputID {
		txID := input.nextAliasOutput.TransactionID()
		cmi.needPublishTX[txID] = &NeedPublishTX{
			CommitteeAddr:     input.committeeAddr,
			TxID:              txID,
			Tx:                input.transaction,
			BaseAliasOutputID: input.baseAliasOutputID,
			NextAliasOutput:   input.nextAliasOutput,
		}
	}
	//
	// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
	msgs := cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmtLog.NewInputConsensusOutputDone(input.logIndex, input.baseAliasOutputID, input.nextAliasOutput))
	})
	//
	// >     Update AccessNodes.
	newAccessNodes := governance.NewStateAccess(input.nextState).GetAccessNodes()
	if !util.Same(newAccessNodes, cmi.activeAccessNodes) {
		cmi.activeAccessNodesCB(newAccessNodes)
		cmi.activeAccessNodes = newAccessNodes
	}
	//
	return msgs
}

// > UPON Reception of Consensus Output/SKIP:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
func (cmi *chainMgrImpl) handleInputConsensusOutputSkip(input *inputConsensusOutputSkip) gpa.OutMessages {
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmtLog.NewInputConsensusOutputSkip(input.logIndex, input.baseAliasOutputID))
	})
}

// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
func (cmi *chainMgrImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	cmi.log.Debugf("handleInputConsensusTimeout: %+v", input)
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmtLog.NewInputConsensusTimeout(input.logIndex))
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
		cmi.latestActiveCmt = &cli.committeeAddr
		cmi.ensureNeedConsensus(cli, outputUntyped)
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
	}
	cmi.latestActiveCmt = &cli.committeeAddr
	cmi.ensureNeedConsensus(cli, outputUntyped)
	return msgs
}

func (cmi *chainMgrImpl) ensureNeedConsensus(cli *cmtLogInst, outputUntyped gpa.Output) {
	if outputUntyped == nil {
		cmi.needConsensus = nil
		return
	}
	output := outputUntyped.(*cmtLog.Output)
	if cmi.needConsensus != nil && cmi.needConsensus.IsFor(output) {
		// Not changed, keep it.
		return
	}
	committeeAddress := output.GetBaseAliasOutput().GetStateAddress()
	dkShare, err := cmi.dkShareRegistryProvider.LoadDKShare(committeeAddress)
	if err != nil {
		panic(xerrors.Errorf("cannot load DKShare for %v", committeeAddress))
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
func (cmi *chainMgrImpl) StatusString() string {
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
		return cmi.wrapCmtLogMsgs(cli, cli.gpaInstance.Input(cmtLog.NewInputSuspend()))
	}
	return nil
}

func (cmi *chainMgrImpl) withCmtLog(committeeAddr iotago.Ed25519Address, handler func(cl gpa.GPA) gpa.OutMessages) gpa.OutMessages {
	cli, clMsgs, err := cmi.ensureCmtLog(committeeAddr)
	if err != nil {
		cmi.log.Warnf("cannot find committee: %v", committeeAddr)
		return nil
	}
	return gpa.NoMessages().
		AddAll(clMsgs).
		AddAll(cmi.handleCmtLogOutput(cli, handler(cli.gpaInstance)))
}

// NOTE: ErrNotInCommittee
func (cmi *chainMgrImpl) ensureCmtLog(committeeAddr iotago.Ed25519Address) (*cmtLogInst, gpa.OutMessages, error) {
	if cli, ok := cmi.cmtLogs[committeeAddr]; ok {
		return cli, nil, nil
	}
	//
	// Create a committee if not created yet.
	dkShare, err := cmi.dkShareRegistryProvider.LoadDKShare(&committeeAddr)
	if errors.Is(err, tcrypto.ErrDKShareNotFound) {
		return nil, nil, ErrNotInCommittee
	}
	if err != nil {
		return nil, nil, xerrors.Errorf("cannot load DKShare for committeeAddress=%v: %w", committeeAddr, err)
	}

	clInst, err := cmtLog.New(cmi.me, cmi.chainID, dkShare, cmi.cmtLogStore, cmi.nodeIDFromPubKey, cmi.log)
	if err != nil {
		return nil, nil, xerrors.Errorf("cannot create cmtLog for committeeAddress=%v: %w", committeeAddr, err)
	}
	clGPA := clInst.AsGPA()
	cli := &cmtLogInst{
		committeeAddr: committeeAddr,
		gpaInstance:   clGPA,
		pendingMsgs:   []gpa.Message{},
	}
	cmi.cmtLogs[committeeAddr] = cli
	msgs := cmi.handleCmtLogOutput(cli, clGPA.Input(cmtLog.NewInputStart()))
	return cli, msgs, nil
}
