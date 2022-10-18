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
// >     NeedToPublishTX -- Requests to publish TX'es.
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
// >     Clear the TX from the NeedToPublishTX variable.
// >     If result.confirmed = false THEN
// >         Forward it to ChainMgr; HandleCmtLogOutput.
// >     ELSE
// >         NOP // AO has to be received as Confirmed AO.
// > UPON Reception of Consensus Output:
// >     IF ConsensusOutput.BaseAO == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedToPublishTX
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// >     Update AccessNodes.
// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog; HandleCmtLogOutput.
// > UPON Reception of CmtLog.NextLI message:
// >     Forward it to the corresponding CmtLog; HandleCmtLogOutput.
// >
// > PROCEDURE HandleCmtLogOutput:
// >     IF the committee don't match the LatestActiveAO THEN
// >         RETURN
// >     IF output.NeedConsensus != NeedConsensus THEN
// >         Set NeedConsensus <- output.NeedConsensus
// >         Set LatestActiveAO <- output.NeedConsensus
package chainMgr

import (
	"errors"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

var ErrNotInCommittee = errors.New("ErrNotInCommittee")

type CommitteeID = string // AliasOutput().StateController().Key()

func CommitteeIDFromAddress(committeeAddress iotago.Address) CommitteeID {
	return committeeAddress.Key()
}

type Output struct {
	cmi *chainMgrImpl
}

func (o *Output) LatestActiveAliasOutput() *isc.AliasOutputWithID    { return o.cmi.latestActiveAO }
func (o *Output) LatestConfirmedAliasOutput() *isc.AliasOutputWithID { return o.cmi.latestConfirmedAO }
func (o *Output) ActiveAccessNodes() []*cryptolib.PublicKey          { return o.cmi.activeAccessNodes }
func (o *Output) NeedConsensus() *NeedConsensus                      { return o.cmi.needConsensus }
func (o *Output) NeedPostTXes() []NeedPostedTX                       { return o.cmi.needPostTXes }

type NeedConsensus struct {
	CommitteeID     CommitteeID
	LogIndex        journal.LogIndex
	DKShare         tcrypto.DKShare
	BaseAliasOutput *isc.AliasOutputWithID
}

type NeedPostedTX struct {
	CommitteeID CommitteeID
	TxID        iotago.TransactionID
	Tx          *iotago.Transaction
}

type ChainMgr interface {
	AsGPA() gpa.GPA
}

type chainMgrImpl struct {
	chainID           isc.ChainID                      // This instance is responsible for this chain.
	cmtLogs           map[CommitteeID]gpa.GPA          // TODO: ...
	latestActiveAO    *isc.AliasOutputWithID           // The latest AO we are building upon.
	latestConfirmedAO *isc.AliasOutputWithID           // The latest confirmed AO (follows Active AO).
	activeAccessNodes []*cryptolib.PublicKey           // All the nodes authorized for being access nodes (for the ActiveAO).
	needConsensus     *NeedConsensus                   // Query for a consensus.
	needPostTXes      []NeedPostedTX                   // Query to post TXes.
	dkReg             registry.DKShareRegistryProvider // TODO: What ir DKShare is stored after some AO is received?
	output            *Output
	asGPA             gpa.GPA
	me                gpa.NodeID
	nodeIDFromPubKey  func(pubKey *cryptolib.PublicKey) gpa.NodeID
	log               *logger.Logger
}

var (
	_ gpa.GPA  = &chainMgrImpl{}
	_ ChainMgr = &chainMgrImpl{}
)

func New(
	me gpa.NodeID,
	chainID isc.ChainID,
	dkReg registry.DKShareRegistryProvider,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	log *logger.Logger,
) (ChainMgr, error) {
	cmi := &chainMgrImpl{
		chainID:          chainID,
		cmtLogs:          map[CommitteeID]gpa.GPA{},
		dkReg:            dkReg,
		me:               me,
		nodeIDFromPubKey: nodeIDFromPubKey,
		log:              log,
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
// >     Pass it to the corresponding CmtLog.
// >     Send Suspend to all the other CmtLogs.
func (cmi *chainMgrImpl) handleInputAliasOutputConfirmed(input *inputAliasOutputConfirmed) gpa.OutMessages {
	cmi.log.Debugf("handleInputAliasOutputConfirmed: %v", input.aliasOutput)
	cmi.latestConfirmedAO = input.aliasOutput // The latest AO is always the most correct.
	msgs := gpa.NoMessages()
	committeeAddress := input.aliasOutput.GetAliasOutput().StateController()
	committeeLog, committeeID, cmtMsgs, err := cmi.ensureCmtLog(committeeAddress)
	msgs.AddAll(cmtMsgs)
	if errors.Is(err, ErrNotInCommittee) {
		cmi.log.Debugf("This node is not in the committee for aliasOutput: %v", input.aliasOutput)
		return nil
	}
	if err != nil {
		cmi.log.Warnf("Failed to get CmtLog: %v", err)
		return nil
	}
	msgs.AddAll(cmi.handleCmtLogOutput(
		committeeLog, committeeID,
		committeeLog.Message(cmtLog.NewMsgAliasOutputConfirmed(cmi.me, input.aliasOutput)),
	))
	for cid, cl := range cmi.cmtLogs {
		if cid == committeeID {
			continue
		}
		msgs.AddAll(cmi.handleCmtLogOutput(
			cl, cid,
			cl.Message(cmtLog.NewMsgSuspend(cmi.me)),
		))
	}
	return msgs
}

// > UPON Reception of PublishResult:
// >     // TODO: ...
func (cmi *chainMgrImpl) handleInputChainTxPublishResult(input *inputChainTxPublishResult) gpa.OutMessages {
	if input.confirmed {
		// TODO: Delete it from the needed tx pubs
		// TODO: Call the handleMsgAliasOutputConfirmed???
		return nil
	}
	// TODO: Send reject to the appropriate CL.
	return nil // TODO: ...
}

// TODO: Have to be used for pipelining. We have started to publish the TX, we can try to build on it.
//
// > UPON Reception of Consensus Output:
// >     Forward the message to the corresponding CmtLog.
// >     // TODO: Add to TX'es to publish?
// >     // TODO: Clear the request for consensus?
func (cmi *chainMgrImpl) handleInputConsensusOutput(input *inputConsensusOutput) gpa.OutMessages {
	committeeLog, ok := cmi.cmtLogs[input.committeeID]
	if !ok {
		cmi.log.Warnf("Discarding consensus output for unknown committeeID: %+v", input)
		return nil
	}
	return cmi.handleCmtLogOutput( // TODO: Cleanup request for a consensus? Thats in the handler probably?
		committeeLog, input.committeeID,
		committeeLog.Message(cmtLog.NewMsgConsensusOutput(cmi.me, input.logIndex, input.baseAliasOutputID, input.nextAliasOutput)),
	)
}

// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog.
func (cmi *chainMgrImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	committeeLog, ok := cmi.cmtLogs[input.committeeID]
	if !ok {
		cmi.log.Warnf("Dropping msgConsensusTimeout for unknown committeeID: %+v", input)
		return nil
	}
	return cmi.handleCmtLogOutput(
		committeeLog, input.committeeID,
		committeeLog.Message(cmtLog.NewMsgConsensusTimeout(cmi.me, input.logIndex)),
	)
}

// > UPON Reception of CmtLog.NextLI message:
// >     Forward it to the corresponding CmtLog.
func (cmi *chainMgrImpl) handleMsgCmtLog(msg *msgCmtLog) gpa.OutMessages {
	committeeLog, ok := cmi.cmtLogs[msg.committeeID]
	if !ok {
		cmi.log.Warnf("Message for non-existing CmtLog: %+v", msg)
	}
	return cmi.handleCmtLogOutput(
		committeeLog, msg.committeeID,
		committeeLog.Message(msg.wrapped),
	)
}

// NOTE: ErrNotInCommittee
func (cmi *chainMgrImpl) ensureCmtLog(committeeAddress iotago.Address) (gpa.GPA, CommitteeID, gpa.OutMessages, error) {
	committeeID := CommitteeIDFromAddress(committeeAddress)
	if cl, ok := cmi.cmtLogs[committeeID]; ok {
		return cl, committeeID, nil, nil
	}
	//
	// Create a committee if not created yet.
	dkShare, err := cmi.dkReg.LoadDKShare(committeeAddress)
	if errors.Is(err, registry.ErrDKShareNotFound) { // TODO: This interface and error definition should be along with the DKShare.
		return nil, committeeID, nil, ErrNotInCommittee
	}
	if err != nil {
		return nil, committeeID, nil, xerrors.Errorf("cannot load DKShare for committeeAddress=%v: %w", committeeAddress, err)
	}

	clInst, err := cmtLog.New(cmi.me, cmi.chainID, dkShare, nil, cmi.nodeIDFromPubKey, cmi.log) // TODO: Pass Store
	if err != nil {
		return nil, committeeID, nil, xerrors.Errorf("cannot create cmtLog for committeeAddress=%v: %w", committeeAddress, err)
	}
	cl := clInst.AsGPA()
	cmi.cmtLogs[committeeID] = cl
	msgs := cmi.handleCmtLogOutput(
		cl, committeeID, cl.Input(nil),
	)
	return cl, committeeID, msgs, nil
}

func (cmi *chainMgrImpl) handleCmtLogOutput(committeeLog gpa.GPA, committeeID CommitteeID, outMsgs gpa.OutMessages) gpa.OutMessages {
	// TODO: wrap msgs, handle outputs ...
	return outMsgs
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Output() gpa.Output {
	return cmi.output
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) StatusString() string {
	return "{}" // TODO: Implement.
}
