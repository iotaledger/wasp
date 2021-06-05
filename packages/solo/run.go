// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/stretchr/testify/require"
)

func (ch *Chain) runRequestsSync(reqs []coretypes.Request, trace string) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	ch.reqCounter.Add(int32(len(reqs)))
	ch.mempool.ReceiveRequests(reqs...)
	ch.mempool.WaitAllRequestsIn()

	//ready := ch.mempool.ReadyNow(ch.Env.LogicalTime())
	//require.EqualValues(ch.Env.T, len(reqs), len(ready))
	return ch.runRequestsNolock(reqs, trace)
}

func (ch *Chain) runRequestsNolock(reqs []coretypes.Request, trace string) (dict.Dict, error) {
	ch.Log.Debugf("runRequestsSync ('%s')", trace)

	for _, r := range reqs {
		_, solidArgs := r.Params()
		require.True(ch.Env.T, solidArgs)
	}
	task := &vm.VMTask{
		Processors:         ch.proc,
		ChainInput:         ch.GetChainOutput(),
		Requests:           reqs,
		Timestamp:          ch.Env.LogicalTime(),
		VirtualState:       ch.State,
		Entropy:            hashing.RandomHash(nil),
		ValidatorFeeTarget: ch.ValidatorFeeTarget,
		Log:                ch.Log,
	}
	var err error
	var callRes dict.Dict
	var callErr error
	// state baseline always valid in Solo
	task.SolidStateBaseline = ch.GlobalSync.GetSolidIndexBaseline()
	task.OnFinish = func(callResult dict.Dict, callError error, err error) {
		require.NoError(ch.Env.T, err)
		callRes = callResult
		callErr = callError
		ch.reqCounter.Add(int32(-len(task.Requests)))
	}

	ch.Env.vmRunner.Run(task)

	ch.Env.AdvanceClockBy(time.Duration(len(task.Requests)+1) * time.Nanosecond)

	inputs, err := ch.Env.utxoDB.CollectUnspentOutputsFromInputs(task.ResultTransactionEssence)
	require.NoError(ch.Env.T, err)
	unlockBlocks, err := utxoutil.UnlockInputsWithED25519KeyPairs(inputs, task.ResultTransactionEssence, ch.StateControllerKeyPair)
	require.NoError(ch.Env.T, err)

	tx := ledgerstate.NewTransaction(task.ResultTransactionEssence, unlockBlocks)
	ch.settleStateTransition(tx, coretypes.TakeRequestIDs(reqs...))

	return callRes, callErr
}

func (ch *Chain) settleStateTransition(stateTx *ledgerstate.Transaction, reqids []coretypes.RequestID) {
	err := ch.Env.AddToLedger(stateTx)
	require.NoError(ch.Env.T, err)

	stateOutput, err := utxoutil.GetSingleChainedAliasOutput(stateTx)
	require.NoError(ch.Env.T, err)

	// saving block just to check consistency. Otherwise, saved blocks are not used in Solo
	block, err := ch.State.ExtractBlock()
	require.NoError(ch.Env.T, err)
	require.NotNil(ch.Env.T, block)
	block.SetApprovingOutputID(stateOutput.ID())

	err = ch.State.Commit(block)
	require.NoError(ch.Env.T, err)

	blockBack, err := state.LoadBlock(ch.Env.dbmanager.GetKVStore(&ch.ChainID), ch.State.BlockIndex())
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, bytes.Equal(block.Bytes(), blockBack.Bytes()))
	require.EqualValues(ch.Env.T, stateOutput.ID(), blockBack.ApprovingOutputID())

	chain.PublishStateTransition(ch.State, stateOutput, reqids)

	ch.Log.Infof("state transition --> #%d. Requests in the block: %d. Outputs: %d",
		ch.State.BlockIndex(), len(reqids), len(stateTx.Essence().Outputs()))
	ch.Log.Debugf("Batch processed: %s", batchShortStr(reqids))

	ch.mempool.RemoveRequests(reqids...)

	go ch.Env.EnqueueRequests(stateTx)
	ch.Env.ClockStep()
}

func batchShortStr(reqIds []coretypes.RequestID) string {
	ret := make([]string, len(reqIds))
	for i, r := range reqIds {
		ret[i] = r.Short()
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ","))
}
