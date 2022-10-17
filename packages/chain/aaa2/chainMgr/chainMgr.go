// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// TODO: Cleanup the committees not used for a long time.

// This package implements a protocol for running a chain in a node. Its main responsibilities:
//   - Decide, which branch is the correct one.
//   - Maintain a set of committee logs (1 for each committee this node participates in).
//   - Maintain a set of consensus instances (one of them is the current one).
//   - Supervise the Mempool and StateMgr.
//   - Handle messages from the NodeConn (AO confirmed / rejected, Request received).
//   - Posting StateTX to NodeConn.
//   - ...
//
// TODO: Manage the access nodes (per node and from the governance contract).
// TODO: Output contains consensus needed to run,
// TODO: Pass AO to the cmtLog, update the consensus accordingly.
// TODO: NodeConn has to return AO chains (all in the milestone at once).
// TODO: Don't persist TX'es to post. In-memory is enough.
// TODO: Save Block to StateMgr from the Consensus.
// TODO: Where the OriginState should be created?
// TODO: We have to track the main "branch" here, and provide access to the HEAD.
//
// TODO: Pass MSG to CmtLog: AO Confirmed
// TODO: Pass MSG to CmtLog: AO Rejected
// TODO: Pass MSG to CmtLog: AO Consensus Done
// TODO: Pass MSG to CmtLog: AO Consensus Timeout
// TODO: Pass MSG to CmtLog: AO Suspend
// TODO: Pass MSG to CmtLog: AO TimerTick
// TODO: Wrap CmtLog out messages (NextLI).
//
// > UPON Reception of Confirmed AO:
// >     Pass it to the corresponding CmtLog.
// >     Send Suspend to all the other CmtLogs.
// > UPON Reception of PublishResult:
// >     // TODO: ...
// > UPON Reception of Consensus Output:
// >     Forward the message to the corresponding CmtLog.
// >     // TODO: Add to TX'es to publish?
// >     // TODO: Clear the request for consensus?
// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog.
// > UPON Reception of CmtLog.NextLI message:
// >     Forward it to the corresponding CmtLog.
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
	NeedConsensus *NeedConsensus
	NeedPostTXes  []NeedPostedTX
}
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
	chainID          isc.ChainID                      // This instance is responsible for this chain.
	cmtLogs          map[CommitteeID]gpa.GPA          // TODO: ...
	dkReg            registry.DKShareRegistryProvider // TODO: What ir DKShare is stored after some AO is received?
	output           *Output
	asGPA            gpa.GPA
	me               gpa.NodeID
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID
	log              *logger.Logger
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
		output:           &Output{NeedConsensus: nil, NeedPostTXes: nil},
		me:               me,
		nodeIDFromPubKey: nodeIDFromPubKey,
		log:              log,
	}
	cmi.asGPA = gpa.NewOwnHandler(me, cmi)
	return cmi, nil
}

// Implements the CmtLog interface.
func (cmi *chainMgrImpl) AsGPA() gpa.GPA {
	return cmi.asGPA
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Input(input gpa.Input) gpa.OutMessages {
	panic(xerrors.Errorf("should not be used"))
}

// Implements the gpa.GPA interface.
func (cmi *chainMgrImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msg := msg.(type) {
	case *msgAliasOutputConfirmed:
		return cmi.handleMsgAliasOutputConfirmed(msg)
	case *msgChainTxPublishResult:
		return cmi.handleMsgChainTxPublishResult(msg)
	case *msgConsensusOutput:
		return cmi.handleMsgConsensusOutput(msg)
	case *msgConsensusTimeout:
		return cmi.handleMsgConsensusTimeout(msg)
	case *msgCmtLog:
		return cmi.handleMsgCmtLog(msg)
	}
	cmi.log.Warnf("dropping unexpected message: %+v", msg)
	return nil
}

// > UPON Reception of Confirmed AO:
// >     Pass it to the corresponding CmtLog.
// >     Send Suspend to all the other CmtLogs.
func (cmi *chainMgrImpl) handleMsgAliasOutputConfirmed(msg *msgAliasOutputConfirmed) gpa.OutMessages {
	cmi.log.Debugf("handleMsgAliasOutputConfirmed: %v", msg.aliasOutput)
	msgs := gpa.NoMessages()
	committeeAddress := msg.aliasOutput.GetAliasOutput().StateController()
	committeeLog, committeeID, cmtMsgs, err := cmi.ensureCmtLog(committeeAddress)
	msgs.AddAll(cmtMsgs)
	if errors.Is(err, ErrNotInCommittee) {
		cmi.log.Debugf("This node is not in the committee for aliasOutput: %v", msg.aliasOutput)
		return nil
	}
	if err != nil {
		cmi.log.Warnf("Failed to get CmtLog: %v", err)
		return nil
	}
	msgs.AddAll(cmi.handleCmtLogOutput(
		committeeLog, committeeID,
		committeeLog.Message(cmtLog.NewMsgAliasOutputConfirmed(cmi.me, msg.aliasOutput)),
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
func (cmi *chainMgrImpl) handleMsgChainTxPublishResult(msg *msgChainTxPublishResult) gpa.OutMessages {
	if msg.confirmed {
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
func (cmi *chainMgrImpl) handleMsgConsensusOutput(msg *msgConsensusOutput) gpa.OutMessages {
	committeeLog, ok := cmi.cmtLogs[msg.committeeID]
	if !ok {
		cmi.log.Warnf("Discarding consensus output for unknown committeeID: %+v", msg)
		return nil
	}
	return cmi.handleCmtLogOutput( // TODO: Cleanup request for a consensus? Thats in the handler probably?
		committeeLog, msg.committeeID,
		committeeLog.Message(cmtLog.NewMsgConsensusOutput(cmi.me, msg.logIndex, msg.baseAliasOutputID, msg.nextAliasOutput)),
	)
}

// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CmtLog.
func (cmi *chainMgrImpl) handleMsgConsensusTimeout(msg *msgConsensusTimeout) gpa.OutMessages {
	committeeLog, ok := cmi.cmtLogs[msg.committeeID]
	if !ok {
		cmi.log.Warnf("Dropping msgConsensusTimeout for unknown committeeID: %+v", msg)
		return nil
	}
	return cmi.handleCmtLogOutput(
		committeeLog, msg.committeeID,
		committeeLog.Message(cmtLog.NewMsgConsensusTimeout(cmi.me, msg.logIndex)),
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
