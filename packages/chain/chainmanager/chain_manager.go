// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// TODO: Cleanup the committees not used for a long time.

// Package chainmanager implements a protocol for running a chain in a node.
// Its main responsibilities:
//   - Track, which branch is the latest/correct one.
//   - Maintain a set of committee logs (1 for each committee this node participates in).
//   - Maintain a set of consensus instances (one of them is the current one).
//   - Supervise the Mempool and StateMgr.
//   - Handle messages from the NodeConn (Anchor confirmed / rejected, Request received).
//   - Posting StateTX to NodeConn.
//
// > VARIABLES:
// >     LatestActiveCommittee -- The latest committee, that was active.
// >        This field will be nil if the node is not part of the committee.
// >        On the resynchronization it will store the previous active committee.
// >     LatestActiveAnchor -- The latest Anchor we are building upon.
// >        Derived, equal to NeedConsensus.BaseAnchor.
// >     LatestConfirmedAnchor -- The latest ConfirmedAnchor from L1.
// >        This one usually follows the LatestAnchor,
// >        but can be published from outside and override the LatestAnchor.
// >     AccessNodes -- The set of access nodes for the current head.
// >        Union of On-Chain access nodes and the nodes permitted by this node.
// >     NeedConsensus -- A request to run consensus.
// >        Always set based on output of the main CommitteeLog.
// >     NeedPublishTX -- Requests to publish TX'es.
// >        - Added upon reception of the Consensus Output,
// >          if it is still in NeedConsensus at the time.
// >        - Removed on PublishResult from the NodeConn.
// >
// > UPON Reception of ConfirmedAnchor:
// >     Set LatestConfirmedAnchor <- ConfirmedAnchor
// >     IF this node is in the committee THEN
// >         Pass it to the corresponding CommitteeLog; HandleCommitteeLogOutput.
// >     ELSE
// >         IF LatestActiveCommittee != nil THEN
// >     	     Send Suspend to Last Active CommitteeLog; HandleCommitteeLogOutput
// >         Set LatestActiveCommittee <- NIL
// >         Set NeedConsensus <- NIL
// > UPON Reception of PublishResult:
// >     Clear the TX from the NeedPublishTX variable.
// >     If result.confirmed = false THEN
// >         Forward it to ChainMgr; HandleCommitteeLogOutput.
// >     ELSE
// >         NOP // Anchor has to be received as ConfirmedAnchor.
// > UPON Reception of Consensus Output/DONE:
// >     IF ConsensusOutput.BaseAnchor == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
// >     Update AccessNodes.
// > UPON Reception of Consensus Output/SKIP:
// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
// > UPON Reception of CommitteeLog.NextLI message:
// >     Forward it to the corresponding CommitteeLog; HandleCommitteeLogOutput.
// >
// > PROCEDURE HandleCommitteeLogOutput(committee):
// >     Wrap out messages.
// >     IF committee == LatestActiveCommittee || LatestActiveCommittee == NIL THEN
// >         Set LatestActiveCommittee <- committee
// >         Set NeedConsensus <- output.NeedConsensus // Can be nil
// >     ELSE
// >         IF output.NeedConsensus == nil THEN
// >             RETURN // No need to change the committee.
// >         IF LatestActiveCommittee != nil THEN
// >             Suspend(LatestActiveCommittee)
// >         Set LatestActiveCommittee <- committee
// >         Set NeedConsensus <- output.NeedConsensus
//
// TODO: Why AM is not notified on the committee nodes after rotation?
package chainmanager

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

var ErrNotInCommittee = errors.New("ErrNotInCommittee")

type Output struct {
	chainMgr *chainMgrImpl
}

func (o *Output) LatestActiveAnchor() *isc.StateAnchor {
	// There is no pipelining possible with the SUI based L1,
	// thus there is no difference between the active and confirmed Anchor.
	return o.chainMgr.latestConfirmedAnchor
}
func (o *Output) LatestConfirmedAnchor() *isc.StateAnchor { return o.chainMgr.latestConfirmedAnchor }
func (o *Output) NeedPublishTX() *NeedPublishTXMap {
	return o.chainMgr.needPublishTX
}

func (o *Output) String() string {
	needPublishTX := "{"
	for txID, needPub := range o.NeedPublishTX().AsMap() {
		needPublishTX += fmt.Sprintf("%s => %v; ", txID.Hex(), needPub)
	}
	needPublishTX += "}"

	return fmt.Sprintf(
		"{m.Output, LatestConfirmedAnchor=%v, NeedConsensus=%v, NeedPublishTX=%s}",
		o.LatestConfirmedAnchor(),
		o.chainMgr.needConsensus,
		needPublishTX,
	)
}

// NeedConsensusKeySize is the required consensus key size in bytes
// We use NeedConsensusKey to address the the instances in a map.
const NeedConsensusKeySize = cryptolib.AddressSize + 4

type NeedConsensusKey [NeedConsensusKeySize]byte

func (nck NeedConsensusKey) String() string {
	return hexutil.Encode(nck[:])
}

type NeedConsensusMap = shrinkingmap.ShrinkingMap[NeedConsensusKey, *NeedConsensus]

func MakeConsensusKey(committeeAddr cryptolib.Address, logIndex committeelog.LogIndex) NeedConsensusKey {
	var buf [NeedConsensusKeySize]byte
	cak := committeeAddr.Key()
	lib := logIndex.Bytes()
	copy(buf[0:cryptolib.AddressSize], cak[0:cryptolib.AddressSize])
	copy(buf[cryptolib.AddressSize:NeedConsensusKeySize], lib[0:4])
	return buf
}

type NeedConsensus struct {
	CommitteeAddr   cryptolib.Address
	DistKeyPart     tcrypto.DistibutedKeyPart
	LogIndex        committeelog.LogIndex
	BaseStateAnchor *isc.StateAnchor
}

func (nc *NeedConsensus) String() string {
	return fmt.Sprintf(
		"{m.NeedConsensus, CommitteeAddr=%v, LogIndex=%v, BaseStateAnchor=%v}",
		nc.CommitteeAddr.String(),
		nc.LogIndex,
		nc.BaseStateAnchor,
	)
}

type NeedPublishTXMap = shrinkingmap.ShrinkingMap[hashing.HashValue, *NeedPublishTX]

type NeedPublishTX struct {
	CommitteeAddr cryptolib.Address
	LogIndex      committeelog.LogIndex
	Tx            *iotasigner.SignedTransaction
	BaseAnchorRef *iotago.ObjectRef // The consumed Anchor object/version.
}

func (npt *NeedPublishTX) String() string {
	return fmt.Sprintf(
		"{m.NeedPublishTX, CommitteeAddr=%v, LogIndex=%v, TX=..., BaseAnchorRef=%v}",
		npt.CommitteeAddr.String(),
		npt.LogIndex,
		npt.BaseAnchorRef,
	)
}

type ChainMgr interface {
	AsGPA() gpa.GPA
}

type committeeLogInst struct {
	committeeAddr cryptolib.Address
	distKeyPart   tcrypto.DistibutedKeyPart
	gpaInstance   gpa.GPA
	pendingMsgs   []gpa.Message
}

type chainMgrImpl struct {
	chainID                     isc.ChainID                                             // This instance is responsible for this chain.
	chainStore                  state.Store                                             // Store of the chain state.
	committeeLogs               map[cryptolib.AddressKey]*committeeLogInst              // All the committee log instances for this chain.
	consensusStateRegistry      committeelog.ConsensusStateRegistry                     // Persistent store for log indexes.
	latestActiveCommittee       *cryptolib.Address                                      // The latest active committee.
	latestConfirmedAnchor       *isc.StateAnchor                                        // The latest confirmed Anchor (follows Active Anchor).
	activeNodesCB               func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey) // All the nodes authorized for being access nodes (for the ActiveAnchor).
	trackActiveStateCB          func(ao *isc.StateAnchor)                               // We will call this to set new Anchor for the active state.
	savePreliminaryBlockCB      func(block state.Block)                                 // We will call this, when a preliminary block matching the tx signatures is received.
	committeeUpdatedCB          func(distKeyPart tcrypto.DistibutedKeyPart)             // Will be called, when a committee changes.
	needConsensus               *NeedConsensusMap                                       // Query for a consensus.
	needConsensusCB             func(upd *NeedConsensusMap)                             // A callback.
	needPublishTX               *NeedPublishTXMap                                       // Query to post TXes.
	needPublishCB               func(upd *NeedPublishTXMap)                             // A callback.
	distKeyPartRegistryProvider registry.DistKeyPartRegistryProvider                    // Source for DistKeyParts.
	accessNodeStateVariable     AccessNodeStateVariable
	output                      *Output
	asGPA                       gpa.GPA
	me                          gpa.NodeID
	nodeIDFromPubKey            func(pubKey *cryptolib.PublicKey) gpa.NodeID
	deriveAnchorByQuorum        bool // Config parameter.
	pipeliningLimit             int  // Config parameter.
	postponeRecoveryMilestones  int  // Config parameter.
	metrics                     *metrics.ChainCommitteeLogMetrics
	log                         log.Logger
}

var (
	_ gpa.GPA  = &chainMgrImpl{}
	_ ChainMgr = &chainMgrImpl{}
)

func New(
	me gpa.NodeID,
	chainID isc.ChainID,
	chainStore state.Store,
	consensusStateRegistry committeelog.ConsensusStateRegistry,
	distKeyPartRegistryProvider registry.DistKeyPartRegistryProvider,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	needConsensusCB func(upd *NeedConsensusMap),
	needPublishCB func(upd *NeedPublishTXMap),
	activeNodesCB func() ([]*cryptolib.PublicKey, []*cryptolib.PublicKey),
	trackActiveStateCB func(ao *isc.StateAnchor),
	savePreliminaryBlockCB func(block state.Block),
	committeeUpdatedCB func(distKeyPart tcrypto.DistibutedKeyPart),
	deriveAnchorByQuorum bool, // TODO: Review, some of them are outdated.
	pipeliningLimit int,
	postponeRecoveryMilestones int,
	metrics *metrics.ChainCommitteeLogMetrics,
	log log.Logger,
) (ChainMgr, error) {
	m := &chainMgrImpl{
		chainID:                     chainID,
		chainStore:                  chainStore,
		committeeLogs:               map[cryptolib.AddressKey]*committeeLogInst{},
		consensusStateRegistry:      consensusStateRegistry,
		activeNodesCB:               activeNodesCB,
		trackActiveStateCB:          trackActiveStateCB,
		savePreliminaryBlockCB:      savePreliminaryBlockCB,
		committeeUpdatedCB:          committeeUpdatedCB,
		needConsensus:               shrinkingmap.New[NeedConsensusKey, *NeedConsensus](),
		needConsensusCB:             needConsensusCB,
		needPublishTX:               shrinkingmap.New[hashing.HashValue, *NeedPublishTX](),
		needPublishCB:               needPublishCB,
		distKeyPartRegistryProvider: distKeyPartRegistryProvider,
		accessNodeStateVariable:     NewAccessNodeStateVariable(chainID, log.NewChildLogger("VAS")),
		me:                          me,
		nodeIDFromPubKey:            nodeIDFromPubKey,
		deriveAnchorByQuorum:        deriveAnchorByQuorum,
		pipeliningLimit:             pipeliningLimit,
		metrics:                     metrics,
		postponeRecoveryMilestones:  postponeRecoveryMilestones,
		log:                         log,
	}
	m.output = &Output{chainMgr: m}
	m.asGPA = gpa.NewOwnHandler(me, m)
	return m, nil
}

// Implements the CommitteeLog interface.
func (m *chainMgrImpl) AsGPA() gpa.GPA {
	return m.asGPA
}

// Implements the gpa.GPA interface.
func (m *chainMgrImpl) Input(input gpa.Input) gpa.OutMessages {
	switch input := input.(type) {
	case *inputAnchorConfirmed:
		return m.handleInputAnchorConfirmed(input)
	case *inputChainTxPublishResult:
		return m.handleInputChainTxPublishResult(input)
	case *inputConsensusOutputDone:
		return m.handleInputConsensusOutputDone(input)
	case *inputConsensusOutputSkip:
		return m.handleInputConsensusOutputSkip(input)
	case *inputConsensusTimeout:
		return m.handleInputConsensusTimeout(input)
	case *inputCanPropose:
		return m.handleInputCanPropose()
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (m *chainMgrImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msg := msg.(type) {
	case *msgCommitteeLog:
		return m.handleMsgCommitteeLog(msg)
	case *msgBlockProduced:
		return m.handleMsgBlockProduced(msg)
	}
	panic(fmt.Errorf("unexpected message %T: %+v", msg, msg))
}

// > UPON Reception of ConfirmedAnchor:
// >     Set LatestConfirmedAnchor <- ConfirmedAnchor
// >     IF this node is in the committee THEN
// >         Pass it to the corresponding CommitteeLog; HandleCommitteeLogOutput(ConfirmedAnchor.Committee).
// >     ELSE
// >         IF LatestActiveCommittee != nil THEN
// >     	     Send Suspend to Last Active CommitteeLog; HandleCommitteeLogOutput(LatestActiveCommittee)
// >         Set LatestActiveCommittee <- NIL
// >         Set NeedConsensus <- NIL
func (m *chainMgrImpl) handleInputAnchorConfirmed(input *inputAnchorConfirmed) gpa.OutMessages {
	m.log.LogDebugf("handleInputAnchorConfirmed: %+v", input)
	//
	// >     Set LatestConfirmedAnchor <- ConfirmedAnchor
	accessNodeStateTip, accessNodeStateUpdated := m.accessNodeStateVariable.BlockConfirmed(input.anchor)
	m.latestConfirmedAnchor = input.anchor
	msgs := gpa.NoMessages()
	committeeLog, err := m.ensureCommitteeLog(*input.stateController) // TODO: input.stateController.Key()
	if errors.Is(err, ErrNotInCommittee) {
		// >     IF this node is in the committee THEN ... ELSE
		// >         IF LatestActiveCommittee != nil THEN
		// >     	     Send Suspend to Last Active CommitteeLog; HandleCommitteeLogOutput(LatestActiveCommittee)
		// >         Set LatestActiveCommittee <- NIL
		// >         Set NeedConsensus <- NIL
		if m.latestActiveCommittee != nil {
			msgs.AddAll(m.suspendCommittee(m.latestActiveCommittee))
			m.committeeUpdatedCB(nil)
			m.latestActiveCommittee = nil
		}
		m.needConsensus.Clear()
		if accessNodeStateUpdated && accessNodeStateTip != nil {
			m.log.LogDebugf("⊢ going to track %v as an access node on confirmed block.", accessNodeStateTip)
			m.trackActiveStateCB(accessNodeStateTip)
		}
		m.log.LogDebugf("This node is not in the committee for anchor: %v", input.anchor)
		return msgs
	}
	if err != nil {
		m.log.LogWarnf("Failed to get CommitteeLog: %v", err)
		return msgs
	}
	// >     IF this node is in the committee THEN
	// >         Pass it to the corresponding CommitteeLog; HandleCommitteeLogOutput.
	msgs.AddAll(m.handleCommitteeLogOutput(
		committeeLog,
		committeeLog.gpaInstance.Input(committeelog.NewInputAnchorConfirmed(input.anchor)),
	))
	return msgs
}

// > UPON Reception of PublishResult:
// >     Clear the TX from the NeedPublishTX variable.
// >     If result.confirmed = false THEN
// >         Forward it to ChainMgr; HandleCommitteeLogOutput.
// >     ELSE
// >         NOP // Anchor has to be received as Confirmed Anchor.
func (m *chainMgrImpl) handleInputChainTxPublishResult(input *inputChainTxPublishResult) gpa.OutMessages {
	m.log.LogDebugf("handleInputChainTxPublishResult: %+v", input)
	// >     Clear the TX from the NeedPublishTX variable.
	if m.needPublishTX.Has(input.txDigest.HashValue()) {
		m.needPublishTX.Delete(input.txDigest.HashValue())
		m.needPublishCB(m.needPublishTX)
	}
	if input.confirmed {
		// >     If result.confirmed = false THEN ... ELSE
		// >         NOP // Anchor has to be received as Confirmed Anchor. // TODO: Not true, anymore.
		return m.withCommitteeLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
			return cl.Input(committeelog.NewInputConsensusOutputConfirmed(input.anchor, input.logIndex))
		})
	}
	// >     If result.confirmed = false THEN
	// >         Forward it to ChainMgr; HandleCommitteeLogOutput.
	return m.withCommitteeLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(committeelog.NewInputConsensusOutputRejected(input.anchor, input.logIndex))
	})
}

// > UPON Reception of Consensus Output/DONE:
// >     IF ConsensusOutput.BaseAnchor == NeedConsensus THEN
// >         Add ConsensusOutput.TX to NeedPublishTX
// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
// >     Update AccessNodes.
func (m *chainMgrImpl) handleInputConsensusOutputDone(input *inputConsensusOutputDone) gpa.OutMessages {
	m.log.LogDebugf("handleInputConsensusOutputDone: %+v", input)
	msgs := gpa.NoMessages()

	baseAnchorRef := input.consensusResult.Transaction.FindInputByID(m.chainID.AsObjectID())
	if baseAnchorRef == nil {
		panic("produced tx is not consuming the anchor")
	}

	// >     IF ConsensusOutput.BaseAnchor == NeedConsensus THEN
	// >         Add ConsensusOutput.TX to NeedPublishTX
	if true { // TODO: Reconsider this condition. Several recent consensus instances should be published, if we run consensus instances in parallel.
		txDigest := lo.Must(input.consensusResult.Transaction.Digest())
		if !m.needPublishTX.Has(txDigest.HashValue()) && input.consensusResult.Block != nil {
			// Inform the access nodes on new block produced.
			block := input.consensusResult.Block
			activeAccessNodes, activeCommitteeNodes := m.activeNodesCB()
			m.log.LogDebugf(
				"Sending MsgBlockProduced (stateIndex=%v, l1Commitment=%v, txDigest=%s) to access nodes: %v except committeeNodes %v",
				block.StateIndex(), block.L1Commitment(), txDigest, activeAccessNodes, activeCommitteeNodes,
			)
			for i := range activeAccessNodes {
				if lo.Contains(activeCommitteeNodes, activeAccessNodes[i]) {
					continue
				}
				msgs.Add(NewMsgBlockProduced(m.nodeIDFromPubKey(activeAccessNodes[i]), input.consensusResult.Transaction, block))
			}
		}
		if !m.needPublishTX.Has(txDigest.HashValue()) {
			m.needPublishTX.Set(txDigest.HashValue(), &NeedPublishTX{
				CommitteeAddr: input.committeeAddr,
				LogIndex:      input.logIndex,
				Tx:            input.consensusResult.Transaction,
				BaseAnchorRef: baseAnchorRef,
			})
			m.needPublishCB(m.needPublishTX)
		}
	}
	//
	// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
	//
	// TODO: This event is not needed anymore.
	// msgs.AddAll(m.withCommitteeLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
	// 	return cl.Input(committeelog.NewInputConsensusOutputDone(input.logIndex, input.proposedBaseAnchor, input.consensusResult))
	// }))
	return msgs
}

// > UPON Reception of Consensus Output/SKIP:
// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
func (m *chainMgrImpl) handleInputConsensusOutputSkip(input *inputConsensusOutputSkip) gpa.OutMessages {
	return m.withCommitteeLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(committeelog.NewInputConsensusOutputSkip(input.logIndex))
	})
}

// > UPON Reception of Consensus Timeout:
// >     Forward the message to the corresponding CommitteeLog; HandleCommitteeLogOutput.
func (m *chainMgrImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	m.log.LogDebugf("handleInputConsensusTimeout: %+v", input)
	return m.withCommitteeLog(input.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(committeelog.NewInputConsensusTimeout(input.logIndex))
	})
}

func (m *chainMgrImpl) handleInputCanPropose() gpa.OutMessages {
	m.log.LogDebugf("handleInputCanPropose")
	return m.withAllCommitteeLogs(func(cl gpa.GPA) gpa.OutMessages {
		return cl.Input(committeelog.NewInputCanPropose())
	})
}

// > UPON Reception of CommitteeLog.NextLI message:
// >     Forward it to the corresponding CommitteeLog; HandleCommitteeLogOutput.
func (m *chainMgrImpl) handleMsgCommitteeLog(msg *msgCommitteeLog) gpa.OutMessages {
	m.log.LogDebugf("handleMsgCommitteeLog: %+v", msg)
	return m.withCommitteeLog(msg.committeeAddr, func(cl gpa.GPA) gpa.OutMessages {
		return cl.Message(msg.wrapped)
	})
}

func (m *chainMgrImpl) handleMsgBlockProduced(msg *msgBlockProduced) gpa.OutMessages {
	m.log.LogDebugf("handleMsgBlockProduced: %+v", msg)
	accessNodeStateTip, accessNodeStateUpdated, l1Commitment := m.accessNodeStateVariable.BlockProduced(msg.tx)
	//
	// Save the block, if it matches all the signatures by the current committee.
	// This will save us a round-trip to query the block from the sender.
	if l1Commitment != nil {
		if msg.block.L1Commitment().Equals(l1Commitment) {
			m.savePreliminaryBlockCB(msg.block)
		} else {
			m.log.LogWarnf("Received msgBlockProduced, but publishedAnchor.l1Commitment != block.l1Commitment.")
		}
	}
	//
	// Update the active state, if needed.
	if accessNodeStateUpdated && accessNodeStateTip != nil && m.latestActiveCommittee == nil {
		m.log.LogDebugf("⊢ going to track %v as an access node on unconfirmed block.", accessNodeStateTip)
		m.trackActiveStateCB(accessNodeStateTip)
	}
	return nil
}

// > PROCEDURE HandleCommitteeLogOutput(committee):
// >     Wrap out messages.
// >     IF committee == LatestActiveCommittee || LatestActiveCommittee == NIL THEN
// >         Set LatestActiveCommittee <- committee
// >         Set NeedConsensus <- output.NeedConsensus // Can be nil
// >     ELSE
// >         IF output.NeedConsensus == nil THEN
// >             RETURN // No need to change the committee.
// >         IF LatestActiveCommittee != nil THEN
// >             Suspend(LatestActiveCommittee)
// >         Set LatestActiveCommittee <- committee
// >         Set NeedConsensus <- output.NeedConsensus
func (m *chainMgrImpl) handleCommitteeLogOutput(cli *committeeLogInst, cliMsgs gpa.OutMessages) gpa.OutMessages {
	//
	// >     Wrap out messages.
	msgs := gpa.NoMessages()
	msgs.AddAll(m.wrapCommitteeLogMsgs(cli, cliMsgs))
	outputUntyped := cli.gpaInstance.Output()
	// >     IF committee == LatestActiveCommittee || LatestActiveCommittee == NIL THEN
	// >         Set LatestActiveCommittee <- committee
	// >         Set NeedConsensus <- output.NeedConsensus // Can be nil
	if m.latestActiveCommittee == nil || cli.committeeAddr.Equals(m.latestActiveCommittee) {
		m.committeeUpdatedCB(cli.distKeyPart)
		m.ensureNeedConsensus(cli, outputUntyped)
		m.latestActiveCommittee = &cli.committeeAddr
		return msgs
	}
	// >     ELSE
	// >         IF output.NeedConsensus == nil THEN
	// >             RETURN // No need to change the committee.
	// >         IF LatestActiveCommittee != nil THEN
	// >             Suspend(LatestActiveCommittee)
	// >         Set LatestActiveCommittee <- committee
	// >         Set NeedConsensus <- output.NeedConsensus
	if outputUntyped == nil {
		return msgs
	}
	if !m.latestActiveCommittee.Equals(&cli.committeeAddr) {
		msgs.AddAll(m.suspendCommittee(m.latestActiveCommittee))
		m.committeeUpdatedCB(cli.distKeyPart)
		m.latestActiveCommittee = &cli.committeeAddr
	}
	m.ensureNeedConsensus(cli, outputUntyped)
	return msgs
}

func (m *chainMgrImpl) ensureNeedConsensus(cli *committeeLogInst, outputUntyped gpa.Output) {
	wasEmpty := m.needConsensus.IsEmpty()
	if outputUntyped == nil {
		m.needConsensus.Clear()
		if !wasEmpty {
			m.needConsensusCB(m.needConsensus)
		}
		return
	}
	output := outputUntyped.(committeelog.Output)
	// if m.needConsensus != nil && m.needConsensus.IsFor(output) {
	// 	// Not changed, keep it.
	// 	return
	// }
	distKeyPart, err := m.distKeyPartRegistryProvider.LoadDistKeyPart(&cli.committeeAddr)
	if errors.Is(err, tcrypto.ErrDistKeyPartNotFound) {
		// Rotated to other nodes, so we don't need to start the next consensus.
		m.needConsensus.Clear()
		if !wasEmpty {
			m.needConsensusCB(m.needConsensus)
		}
		return
	}
	if err != nil {
		panic(fmt.Errorf("ensureNeedConsensus cannot load DistKeyPart for %v: %w", cli.committeeAddr, err))
	}

	//
	// Add new entries, remove those not needed anymore.
	ids := map[NeedConsensusKey]bool{}
	mod := false
	for li, ao := range output {
		key := MakeConsensusKey(cli.committeeAddr, li)
		ids[key] = true
		if m.needConsensus.Has(key) {
			continue
		}
		mod = true
		m.needConsensus.Set(key, &NeedConsensus{
			CommitteeAddr:   cli.committeeAddr,
			LogIndex:        li,
			DistKeyPart:     distKeyPart,
			BaseStateAnchor: ao,
		})
	}
	m.needConsensus.ForEachKey(func(nck NeedConsensusKey) bool {
		if _, ok := ids[nck]; !ok {
			mod = true
			m.needConsensus.Delete(nck)
		}
		return true
	})
	if mod {
		m.needConsensusCB(m.needConsensus)
	}
}

// Implements the gpa.GPA interface.
func (m *chainMgrImpl) Output() gpa.Output {
	return m.output
}

// Implements the gpa.GPA interface.
func (m *chainMgrImpl) StatusString() string { // TODO: Call it periodically. Show the active committee.
	return "{ChainMgr,...}" // TODO: Add more info.
	// return fmt.Sprintf("{ChainMgr,confirmedAnchor=%v,activeAnchor=%v}",
	// 	m.output.LatestConfirmedAnchor().GetObjectID().String(),
	// 	m.output.LatestActiveAnchor().GetObjectID().String(),
	// )
}

////////////////////////////////////////////////////////////////////////////////
// Helper functions.

func (m *chainMgrImpl) wrapCommitteeLogMsgs(cli *committeeLogInst, outMsgs gpa.OutMessages) gpa.OutMessages {
	wrappedMsgs := gpa.NoMessages()
	outMsgs.MustIterate(func(msg gpa.Message) {
		wrappedMsgs.Add(NewMsgCommitteeLog(cli.committeeAddr, msg))
	})
	return wrappedMsgs
}

func (m *chainMgrImpl) suspendCommittee(committeeAddr *cryptolib.Address) gpa.OutMessages {
	for _, cli := range m.committeeLogs {
		if !cli.committeeAddr.Equals(committeeAddr) {
			continue
		}
		return m.wrapCommitteeLogMsgs(cli, cli.gpaInstance.Input(committeelog.NewInputSuspend()))
	}
	return nil
}

func (m *chainMgrImpl) withCommitteeLog(committeeAddr cryptolib.Address, handler func(cl gpa.GPA) gpa.OutMessages) gpa.OutMessages {
	cli, err := m.ensureCommitteeLog(committeeAddr)
	if err != nil {
		m.log.LogWarnf("cannot find committee: %v", committeeAddr)
		return nil
	}
	return gpa.NoMessages().AddAll(m.handleCommitteeLogOutput(cli, handler(cli.gpaInstance)))
}

func (m *chainMgrImpl) withAllCommitteeLogs(handler func(cl gpa.GPA) gpa.OutMessages) gpa.OutMessages {
	msgs := gpa.NoMessages()
	for _, cli := range m.committeeLogs {
		msgs.AddAll(m.handleCommitteeLogOutput(cli, handler(cli.gpaInstance)))
	}
	return msgs
}

// NOTE: ErrNotInCommittee
func (m *chainMgrImpl) ensureCommitteeLog(committeeAddr cryptolib.Address) (*committeeLogInst, error) {
	if cli, ok := m.committeeLogs[committeeAddr.Key()]; ok {
		return cli, nil
	}
	//
	// Create a committee if not created yet.
	distKeyPart, err := m.distKeyPartRegistryProvider.LoadDistKeyPart(&committeeAddr)
	if errors.Is(err, tcrypto.ErrDistKeyPartNotFound) {
		return nil, ErrNotInCommittee
	}
	if err != nil {
		return nil, fmt.Errorf("ensureCommitteeLog cannot load DistKeyPart for committeeAddress=%v: %w", committeeAddr, err)
	}

	clInst, err := committeelog.New(
		m.me,
		m.chainID,
		distKeyPart,
		m.consensusStateRegistry,
		m.nodeIDFromPubKey,
		m.deriveAnchorByQuorum,
		m.pipeliningLimit,
		m.metrics,
		m.log.NewChildLogger(fmt.Sprintf("CL-%v", distKeyPart.GetSharedPublic().AsAddress().String()[:10])),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create committeeLog for committeeAddress=%v: %w", committeeAddr, err)
	}
	clGPA := clInst.AsGPA()
	cli := &committeeLogInst{
		committeeAddr: committeeAddr,
		distKeyPart:   distKeyPart,
		gpaInstance:   clGPA,
		pendingMsgs:   []gpa.Message{},
	}
	m.committeeLogs[committeeAddr.Key()] = cli
	return cli, nil
}
