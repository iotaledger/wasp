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
// >     LatestActiveAO -- The current head we are working on.
// >        The latest received confirmed AO,
// >        OR the output of the main CmtLog.
// >        // Differs from NeedConsensus on access nodes.
// >     LatestConfirmedAO -- The latest confirmed AO from L1.
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
// > UPON Reception of Confirmed AO:
// >     Set the LatestConfirmedAO variable to the received AO.
// >     IF this node is in the committee THEN
// >         Pass it to the corresponding CmtLog; HandleCmtLogOutput.
// >     ELSE
// >         Set LatestActiveAO <- Confirmed AO
// >         Set NeedConsensus <- NIL
// >     Send Suspend to all the other CmtLogs.
// > UPON Reception of PublishResult:
// >     Clear the TX from the NeedPublishTX variable.
// >     If result.confirmed = false THEN
// >         Forward it to ChainMgr; HandleCmtLogOutput.
// >     ELSE
// >         NOP // AO has to be received as Confirmed AO.
// > UPON Reception of Consensus Output:
// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// >     Update AccessNodes.
// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// > UPON Reception of CmtLog.NextLI message:
// >     Forward it to the corresponding CmtLog; HandleCmtLogOutput.
// >
// > PROCEDURE HandleCmtLogOutput:
// >     IF LatestActiveAO != NIL && the committee don't match the LatestActiveAO THEN
// >         RETURN
// >     IF output.NeedConsensus != NeedConsensus THEN
// >         Set NeedConsensus <- output.NeedConsensus
// >         Set LatestActiveAO <- output.NeedConsensus
package chainMgr

import (
	"errors"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

var ErrNotInCommittee = errors.New("ErrNotInCommittee")

type Output struct {
	cmi *chainMgrImpl
}

func (o *Output) LatestActiveAliasOutput() *isc.AliasOutputWithID        { return o.cmi.latestActiveAO }
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
	latestActiveAO          *isc.AliasOutputWithID                  // The latest AO we are building upon.
	latestConfirmedAO       *isc.AliasOutputWithID                  // The latest confirmed AO (follows Active AO).
	activeAccessNodes       []*cryptolib.PublicKey                  // All the nodes authorized for being access nodes (for the ActiveAO).
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
	log *logger.Logger,
) (ChainMgr, error) {
	cmi := &chainMgrImpl{
		chainID:                 chainID,
		cmtLogs:                 map[iotago.Ed25519Address]*cmtLogInst{},
		cmtLogStore:             cmtLogStore,
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
	case *inputConsensusOutput:
		return cmi.handleInputConsensusOutput(input)
	case *inputConsensusTimeout:
		return cmi.handleInputConsensusTimeout(input)
	}
	panic(xerrors.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Message(msg gpa.Message) gpa.OutMessages {
	msgCL, ok := msg.(*msgCmtLog)
	if !ok {
		panic(xerrors.Errorf("unexpected message %T: %+v", msg, msg))
	}
	return cmi.handleMsgCmtLog(msgCL)
}

// > UPON Reception of Confirmed AO:
// >     Set the LatestConfirmedAO variable to the received AO.
// >     IF this node is in the committee THEN
// >         Pass it to the corresponding CmtLog; HandleCmtLogOutput.
// >     ELSE
// >         Set LatestActiveAO <- Confirmed AO
// >         Set NeedConsensus <- NIL
// >     Send Suspend to all the other CmtLogs.
func (cmi *chainMgrImpl) handleInputAliasOutputConfirmed(input *inputAliasOutputConfirmed) gpa.OutMessages {
	cmi.log.Debugf("handleInputAliasOutputConfirmed: %+v", input)
	//
	// >     Set the LatestConfirmedAO variable to the received AO.
	cmi.latestConfirmedAO = input.aliasOutput
	msgs := gpa.NoMessages()
	committeeAddr := input.aliasOutput.GetAliasOutput().StateController().(*iotago.Ed25519Address)
	committeeLog, cmtMsgs, err := cmi.ensureCmtLog(*committeeAddr)
	msgs.AddAll(cmtMsgs)
	if errors.Is(err, ErrNotInCommittee) {
		// >     IF this node is in the committee THEN ... ELSE
		// >         Set LatestActiveAO <- Confirmed AO
		// >         Set NeedConsensus <- NIL
		cmi.latestActiveAO = input.aliasOutput
		cmi.needConsensus = nil
		cmi.log.Debugf("This node is not in the committee for aliasOutput: %v", input.aliasOutput)
		//
		// >     Send Suspend to all (the other) CmtLogs.
		return msgs.AddAll(cmi.suspendAllExcept(nil)) // All.
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
	//
	// >     Send Suspend to all the other CmtLogs.
	msgs.AddAll(cmi.suspendAllExcept(committeeAddr))
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

// > UPON Reception of Consensus Output:
// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// >     Update AccessNodes.
func (cmi *chainMgrImpl) handleInputConsensusOutput(input *inputConsensusOutput) gpa.OutMessages {
	cmi.log.Debugf("handleInputConsensusOutput: %+v", input)
	// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
	// >         Add ConsensusOutput.TX to NeedPublishTX
	if cmi.needConsensus.BaseAliasOutput.ID().Equals(input.baseAliasOutputID.UTXOInput()) {
		txID := input.nextAliasOutput.ID().TransactionID
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
	return cmi.withCmtLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(cmtLog.NewInputConsensusOutput(input.logIndex, input.baseAliasOutputID, input.nextAliasOutput))
	})

	// TODO:
	// >     Update AccessNodes.
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

// Wrap out messages and handle the output as in:
//
// > PROCEDURE HandleCmtLogOutput:
// >     IF LatestActiveAO != NIL && the committee don't match the LatestActiveAO THEN
// >         RETURN
// >     IF output.NeedConsensus != NeedConsensus THEN
// >         Set NeedConsensus <- output.NeedConsensus
// >         Set LatestActiveAO <- output.NeedConsensus
func (cmi *chainMgrImpl) handleCmtLogOutput(cli *cmtLogInst, outMsgs gpa.OutMessages) gpa.OutMessages {
	wrappedMsgs := gpa.NoMessages()
	outMsgs.MustIterate(func(msg gpa.Message) {
		wrappedMsgs.Add(NewMsgCmtLog(cli.committeeAddr, msg))
	})
	outputUntyped := cli.gpaInstance.Output()
	if cmi.latestActiveAO != nil && !cli.committeeAddr.Equal(cmi.latestActiveAO.GetStateAddress()) {
		// >     IF LatestActiveAO != NIL && the committee don't match the LatestActiveAO THEN
		// >         RETURN
		if outputUntyped != nil { // Just an assertion.
			panic(xerrors.Errorf("expecting nil output from non-main cmtLog, but got %v -> %+v", cli.committeeAddr, outputUntyped))
		}
	}
	// >     IF output.NeedConsensus != NeedConsensus THEN
	// >         Set NeedConsensus <- output.NeedConsensus
	// >         Set LatestActiveAO <- output.NeedConsensus
	if outputUntyped == nil {
		cmi.needConsensus = nil
		// TODO: Reset the LatestActiveAO?
		return wrappedMsgs
	}
	output := outputUntyped.(*cmtLog.Output)
	if cmi.needConsensus != nil && cmi.needConsensus.IsFor(output) {
		// Not changed, keep it.
		return wrappedMsgs
	}
	committeeAddress := output.GetBaseAliasOutput().GetStateAddress()
	dkShare, err := cmi.dkShareRegistryProvider.LoadDKShare(committeeAddress)
	if err != nil {
		panic(xerrors.Errorf("cannot load DKShare for %v", committeeAddress))
	}
	cmi.latestActiveAO = output.GetBaseAliasOutput() // TODO:????
	cmi.needConsensus = &NeedConsensus{
		CommitteeAddr:   cli.committeeAddr,
		LogIndex:        output.GetLogIndex(),
		DKShare:         dkShare,
		BaseAliasOutput: output.GetBaseAliasOutput(),
	}
	return wrappedMsgs
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Output() gpa.Output {
	return cmi.output
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) StatusString() string {
	return fmt.Sprintf("{ChainMgr,confirmedAO=%v,activeAO=%v}", cmi.latestConfirmedAO.String(), cmi.latestActiveAO.String())
}

////////////////////////////////////////////////////////////////////////////////
// Helper functions.

func (cmi *chainMgrImpl) suspendAllExcept(committeeAddr *iotago.Ed25519Address) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for ca, cli := range cmi.cmtLogs {
		if ca.Equal(committeeAddr) {
			continue
		}
		msgs.AddAll(cmi.handleCmtLogOutput(
			cli,
			cli.gpaInstance.Input(cmtLog.NewInputSuspend()),
		))
	}
	return msgs
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
