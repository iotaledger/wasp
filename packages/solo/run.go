// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"time"
)

func (ch *Chain) runBatch(batch []coretypes.Request, trace string) (dict.Dict, error) {
	ch.Log.Debugf("runBatch ('%s')", trace)

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	for _, r := range batch {
		_, solidArgs := r.Params()
		require.True(ch.Env.T, solidArgs)
	}
	task := &vm.VMTask{
		Processors:         ch.proc,
		ChainInput:         ch.GetChainOutput(),
		Requests:           batch,
		Timestamp:          ch.Env.LogicalTime(),
		VirtualState:       ch.State.Clone(),
		Entropy:            hashing.RandomHash(nil),
		ValidatorFeeTarget: ch.ValidatorFeeTarget,
		Log:                ch.Log,
	}
	var err error
	var wg sync.WaitGroup
	var callRes dict.Dict
	var callErr error
	task.OnFinish = func(callResult dict.Dict, callError error, err error) {
		require.NoError(ch.Env.T, err)
		callRes = callResult
		callErr = callError
		ch.reqCounter.Add(int32(-len(task.Requests)))
		wg.Done()
	}

	wg.Add(1)
	runvm.MustRunVMTaskAsync(task)
	require.NoError(ch.Env.T, err)
	wg.Wait()

	ch.Env.AdvanceClockBy(time.Duration(len(task.Requests)+1) * time.Nanosecond)

	inputs, err := ch.Env.utxoDB.CollectUnspentOutputsFromInputs(task.ResultTransaction)
	require.NoError(ch.Env.T, err)
	unlockBlocks, err := utxoutil.UnlockInputsWithED25519KeyPairs(inputs, task.ResultTransaction, ch.StateControllerKeyPair)
	require.NoError(ch.Env.T, err)

	tx := ledgerstate.NewTransaction(task.ResultTransaction, unlockBlocks)
	ch.settleStateTransition(task.VirtualState, task.ResultBlock, tx)

	return callRes, callErr
}

func (ch *Chain) settleStateTransition(newState state.VirtualState, block state.Block, stateTx *ledgerstate.Transaction) {
	err := ch.Env.AddToLedger(stateTx)
	require.NoError(ch.Env.T, err)

	err = newState.ApplyBlock(block)
	require.NoError(ch.Env.T, err)

	err = newState.CommitToDb(block)
	require.NoError(ch.Env.T, err)

	reqIDs := blocklog.GetRequestIDsForLastBlock(newState)

	ch.mempool.RemoveRequests(reqIDs...)

	ch.State = newState

	ch.Log.Infof("state transition #%d --> #%d. Requests in the block: %d. Outputs: %d",
		ch.State.BlockIndex()-1, ch.State.BlockIndex(), len(reqIDs), len(stateTx.Essence().Outputs()))
	ch.Log.Debugf("Batch processed: %s", batchShortStr(reqIDs))

	ch.Env.EnqueueRequests(stateTx)
	ch.Env.ClockStep()
}

func batchShortStr(reqIds []coretypes.RequestID) string {
	ret := make([]string, len(reqIds))
	for i, r := range reqIds {
		ret[i] = r.Short()
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ","))
}
