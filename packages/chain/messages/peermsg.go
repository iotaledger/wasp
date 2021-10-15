// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"bytes"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

// Message types for the committee communications.
const (
	MsgGetBlock = 1 + peering.FirstUserMsgCode + iota
	MsgBlock
	MsgSignedResult
	MsgSignedResultAck
	MsgOffLedgerRequest
	MsgMissingRequestIDs
	MsgMissingRequest
	MsgRequestAck
)

type TimerTick int

var _ pipe.Hashable = TimerTick(0)

func (ttT TimerTick) GetHash() interface{} {
	return ttT
}

func (ttT TimerTick) Equals(elem interface{}) bool {
	other, ok := elem.(TimerTick)
	if !ok {
		return false
	}
	return ttT == other
}

type SignedResultMsg struct {
	SenderIndex  uint16
	ChainInputID ledgerstate.OutputID
	EssenceHash  hashing.HashValue
	SigShare     tbls.SigShare
}

type SignedResultAckMsg struct {
	SenderIndex  uint16
	ChainInputID ledgerstate.OutputID
	EssenceHash  hashing.HashValue
}

// GetBlockMsg StateManager queries specific block data from another peer (access node)
type GetBlockMsg struct {
	SenderNetID string
	BlockIndex  uint32
}

// BlockMsg StateManager in response to GetBlockMsg sends block data to the querying node's StateManager
type BlockMsg struct {
	SenderNetID string
	BlockBytes  []byte
}

// DismissChainMsg sent by component to the chain core in case of major setback
type DismissChainMsg struct {
	Reason string
}

var _ pipe.Hashable = &DismissChainMsg{}

func (dcmT *DismissChainMsg) GetHash() interface{} {
	return dcmT.Reason
}

func (dcmT *DismissChainMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*DismissChainMsg)
	if !ok {
		return false
	}
	return dcmT.Reason == other.Reason
}

// StateTransitionMsg Notifies chain about changed state
type StateTransitionMsg struct {
	// new variable state
	State state.VirtualStateAccess
	// corresponding state transaction
	StateOutput *ledgerstate.AliasOutput
	//
	StateTimestamp time.Time
}

var _ pipe.Hashable = &StateTransitionMsg{}

func (stmT *StateTransitionMsg) GetHash() interface{} {
	return struct {
		// TODO: move state fields to state.VirtualStateAccess?
		stateBlockIndex   uint32
		statePreviousHash hashing.HashValue
		stateCommitment   hashing.HashValue
		outputID          ledgerstate.OutputID
	}{
		stateBlockIndex:   stmT.State.BlockIndex(),
		statePreviousHash: stmT.State.PreviousStateHash(),
		stateCommitment:   stmT.State.StateCommitment(),
		outputID:          stmT.StateOutput.ID(),
	}
}

func (stmT *StateTransitionMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*StateTransitionMsg)
	if !ok {
		return false
	}
	// TODO: move state fields to state.VirtualStateAccess?
	if stmT.State.BlockIndex() != other.State.BlockIndex() {
		return false
	}
	if stmT.State.PreviousStateHash() != other.State.PreviousStateHash() {
		return false
	}
	if stmT.State.StateCommitment() != other.State.StateCommitment() {
		return false
	}
	return stmT.StateOutput.ID() == other.StateOutput.ID()
}

// StateCandidateMsg Consensus sends the finalized next state to StateManager
type StateCandidateMsg struct {
	State             state.VirtualStateAccess
	ApprovingOutputID ledgerstate.OutputID
}

var _ pipe.Hashable = &StateCandidateMsg{}

func (scmT *StateCandidateMsg) GetHash() interface{} {
	return struct {
		stateBlockIndex   uint32
		statePreviousHash hashing.HashValue
		stateCommitment   hashing.HashValue
		outputID          ledgerstate.OutputID
	}{
		stateBlockIndex:   scmT.State.BlockIndex(),
		statePreviousHash: scmT.State.PreviousStateHash(),
		stateCommitment:   scmT.State.StateCommitment(),
		outputID:          scmT.ApprovingOutputID,
	}
}

func (scmT *StateCandidateMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*StateCandidateMsg)
	if !ok {
		return false
	}
	if scmT.State.BlockIndex() != other.State.BlockIndex() {
		return false
	}
	if scmT.State.PreviousStateHash() != other.State.PreviousStateHash() {
		return false
	}
	if scmT.State.StateCommitment() != other.State.StateCommitment() {
		return false
	}
	return scmT.ApprovingOutputID == other.ApprovingOutputID
}

// VMResultMsg Consensus -> Consensus. VM sends result of async task started by Consensus to itself
type VMResultMsg struct {
	Task *vm.VMTask
}

var _ pipe.Hashable = &VMResultMsg{}

func (vrmT *VMResultMsg) GetHash() interface{} {
	return vrmT.Task.ACSSessionID
}

func (vrmT *VMResultMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*VMResultMsg)
	if !ok {
		return false
	}
	// NOTE: is it enough???
	return vrmT.Task.ACSSessionID == other.Task.ACSSessionID
}

// AsynchronousCommonSubsetMsg
type AsynchronousCommonSubsetMsg struct {
	ProposedBatchesBin [][]byte
	SessionID          uint64
}

var _ pipe.Hashable = &AsynchronousCommonSubsetMsg{}

func (acsmT *AsynchronousCommonSubsetMsg) GetHash() interface{} {
	var batchesHash hashing.HashValue
	for i := 0; i < len(acsmT.ProposedBatchesBin); i++ {
		batchHash := hashing.HashData(acsmT.ProposedBatchesBin[i])
		for j := 0; j < len(batchesHash); j++ {
			batchesHash[j] += batchHash[j]
		}
	}
	return struct {
		batchesHash hashing.HashValue
		sessionID   uint64
	}{
		batchesHash: batchesHash,
		sessionID:   acsmT.SessionID,
	}
}

func (acsmT *AsynchronousCommonSubsetMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*AsynchronousCommonSubsetMsg)
	if !ok {
		return false
	}
	if acsmT.SessionID != other.SessionID {
		return false
	}
	if len(acsmT.ProposedBatchesBin) != len(other.ProposedBatchesBin) {
		return false
	}
	for i := 0; i < len(acsmT.ProposedBatchesBin); i++ {
		if bytes.Equal(acsmT.ProposedBatchesBin[i], other.ProposedBatchesBin[i]) {
			return false
		}
	}
	return true
}

// InclusionStateMsg txstream plugin sends inclusions state of the transaction to ConsensusOld
type InclusionStateMsg struct {
	TxID  ledgerstate.TransactionID
	State ledgerstate.InclusionState
}

var _ pipe.Hashable = &InclusionStateMsg{}

func (ismT *InclusionStateMsg) GetHash() interface{} {
	return *ismT
}

func (ismT *InclusionStateMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*InclusionStateMsg)
	if !ok {
		return false
	}
	if ismT.TxID != other.TxID {
		return false
	}
	return ismT.State == other.State
}

// StateMsg txstream plugin sends the only existing AliasOutput in the chain's address to StateManager
type StateMsg struct {
	ChainOutput *ledgerstate.AliasOutput
	Timestamp   time.Time
}

var _ pipe.Hashable = &StateMsg{}

func (smT *StateMsg) GetHash() interface{} {
	return smT.ChainOutput.ID()
}

func (smT *StateMsg) Equals(elem interface{}) bool {
	other, ok := elem.(*StateMsg)
	if !ok {
		return false
	}
	return smT.ChainOutput.ID() == other.ChainOutput.ID()
}

func (msg *GetBlockMsg) Write(w io.Writer) error {
	return util.WriteUint32(w, msg.BlockIndex)
}

func (msg *GetBlockMsg) Read(r io.Reader) error {
	return util.ReadUint32(r, &msg.BlockIndex)
}

func (msg *BlockMsg) Write(w io.Writer) error {
	return util.WriteBytes32(w, msg.BlockBytes)
}

func (msg *BlockMsg) Read(r io.Reader) error {
	var err error
	if msg.BlockBytes, err = util.ReadBytes32(r); err != nil {
		return err
	}
	return nil
}

func (msg *SignedResultMsg) Write(w io.Writer) error {
	if _, err := w.Write(msg.EssenceHash[:]); err != nil {
		return err
	}
	if err := util.WriteBytes16(w, msg.SigShare); err != nil {
		return err
	}
	if _, err := w.Write(msg.ChainInputID[:]); err != nil {
		return err
	}
	return nil
}

func (msg *SignedResultMsg) Read(r io.Reader) error {
	if err := util.ReadHashValue(r, &msg.EssenceHash); err != nil {
		return err
	}
	var err error
	if msg.SigShare, err = util.ReadBytes16(r); err != nil {
		return err
	}
	if err := util.ReadOutputID(r, &msg.ChainInputID); /* nolint:revive */ err != nil {
		return err
	}
	return nil
}

func (msg *SignedResultAckMsg) Write(w io.Writer) error {
	if _, err := w.Write(msg.EssenceHash[:]); err != nil {
		return err
	}
	if _, err := w.Write(msg.ChainInputID[:]); err != nil {
		return err
	}
	return nil
}

func (msg *SignedResultAckMsg) Read(r io.Reader) error {
	if err := util.ReadHashValue(r, &msg.EssenceHash); err != nil {
		return err
	}
	return util.ReadOutputID(r, &msg.ChainInputID)
}
