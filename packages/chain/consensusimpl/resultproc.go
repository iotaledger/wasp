// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensusimpl

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/xerrors"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

type runCalculationsParams struct {
	requests        []coretypes.Request
	leaderPeerIndex uint16
	accrueFeesTo    coretypes.AgentID
	timestamp       time.Time
}

// runs the VM for requests and posts result to committee's queue
func (op *operator) runCalculationsAsync(par runCalculationsParams) {
	if op.currentState == nil {
		op.log.Debugf("runCalculationsAsync: variable currentState is not known")
		return
	}
	h := op.stateOutput.ID()
	task := &vm.VMTask{
		Processors:         op.chain.Processors(),
		ChainInput:         op.stateOutput,
		Entropy:            hashing.HashData(h[:]),
		ValidatorFeeTarget: par.accrueFeesTo,
		Requests:           par.requests,
		Timestamp:          par.timestamp,
		VirtualState:       op.currentState.Clone(),
		Log:                op.log,
	}
	task.OnFinish = func(_ dict.Dict, _ error, vmError error) {
		if vmError != nil {
			op.log.Errorf("VM task failed: %v", vmError)
			return
		}
		op.chain.ReceiveMessage(&chain.VMResultMsg{
			Task:   task,
			Leader: par.leaderPeerIndex,
		})
	}
	runvm.MustRunVMTaskAsync(task)
}

func (op *operator) sendResultToTheLeader(result *vm.VMTask, leader uint16) {
	op.log.Debugw("sendResultToTheLeader")
	if op.consensusStage != consensusStageSubCalculationsStarted {
		op.log.Debugf("calculation result on SUB dismissed because stage changed from '%s' to '%s'",
			stages[consensusStageSubCalculationsStarted].name, stages[op.consensusStage].name)
		return
	}

	sigShare, err := op.committee.DKShare().SignShare(result.ResultTransaction.Bytes())
	if err != nil {
		op.log.Errorf("error while signing transaction %v", err)
		return
	}

	reqids := make([]coretypes.RequestID, len(result.Requests))
	for i := range reqids {
		reqids[i] = result.Requests[i].ID()
	}

	essenceHash := hashing.HashData(result.ResultTransaction.Bytes())
	batchHash := vm.BatchHash(reqids, result.Timestamp, leader)

	op.log.Debugw("sendResultToTheLeader",
		"leader", leader,
		"batchHash", batchHash.String(),
		"essenceHash", essenceHash.String(),
		"ts", result.Timestamp,
	)

	msgData := util.MustBytes(&chain.SignedHashMsg{
		StateOutputID: op.stateOutput.ID(),
		BatchHash:     batchHash,
		OrigTimestamp: result.Timestamp.UnixNano(),
		EssenceHash:   essenceHash,
		SigShare:      sigShare,
	})

	if err := op.committee.SendMsg(leader, chain.MsgSignedHash, msgData); err != nil {
		op.log.Error(err)
		return
	}
	op.sentResultToLeader = result.ResultTransaction
	op.sentResultToLeaderIndex = leader

	op.setNextConsensusStage(consensusStageSubCalculationsFinished)
}

func (op *operator) saveOwnResult(result *vm.VMTask) {
	if op.consensusStage != consensusStageLeaderCalculationsStarted {
		op.log.Debugf("calculation result on LEADER dismissed because stage changed from '%s' to '%s'",
			stages[consensusStageLeaderCalculationsStarted].name, stages[op.consensusStage].name)
		return
	}
	sigShare, err := op.committee.DKShare().SignShare(result.ResultTransaction.Bytes())
	if err != nil {
		op.log.Errorf("error while signing transaction %v", err)
		return
	}

	reqids := make([]coretypes.RequestID, len(result.Requests))
	for i := range reqids {
		reqids[i] = result.Requests[i].ID()
	}

	bh := vm.BatchHash(reqids, result.Timestamp, op.committee.OwnPeerIndex())
	if bh != op.leaderStatus.batchHash {
		op.log.Panic("bh != op.leaderStatus.batchHash")
	}

	essenceHash := hashing.HashData(result.ResultTransaction.Bytes())
	op.log.Debugw("saveOwnResult",
		"batchHash", bh.String(),
		"ts", result.Timestamp,
		"essenceHash", essenceHash.String(),
	)

	op.leaderStatus.resultTxEssence = result.ResultTransaction
	op.leaderStatus.virtualState = result.VirtualState
	op.leaderStatus.signedResults[op.committee.OwnPeerIndex()] = &signedResult{
		essenceHash: essenceHash,
		sigShare:    sigShare,
	}
	op.setNextConsensusStage(consensusStageLeaderCalculationsFinished)
}

func (op *operator) aggregateSigShares(sigShares [][]byte) (*ledgerstate.Transaction, error) {
	op.log.Debugf("aggregating sig shares..")
	resTx := op.leaderStatus.resultTxEssence

	signatureWithPK, err := op.committee.DKShare().RecoverFullSignature(sigShares, resTx.Bytes())
	if err != nil {
		return nil, err
	}
	sigUnlockBlock := ledgerstate.NewSignatureUnlockBlock(ledgerstate.NewBLSSignature(*signatureWithPK))
	chainInput := ledgerstate.NewUTXOInput(op.stateOutput.ID())
	var indexChainInput = -1
	for i, inp := range resTx.Inputs() {
		if inp.Compare(chainInput) == 0 {
			indexChainInput = i
			break
		}
	}
	if indexChainInput < 0 {
		return nil, xerrors.New("major inconsistency")
	}
	blocks := make([]ledgerstate.UnlockBlock, len(resTx.Inputs()))
	for i := range op.leaderStatus.resultTxEssence.Inputs() {
		if i == indexChainInput {
			blocks[i] = sigUnlockBlock
		} else {
			blocks[i] = ledgerstate.NewAliasUnlockBlock(uint16(indexChainInput))
		}
	}
	return ledgerstate.NewTransaction(resTx, blocks), nil
}
