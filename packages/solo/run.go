// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"time"
)

func (ch *Chain) validateBatch(batch []*sctransaction.Request) {
}

func (ch *Chain) runBatch(batch []*sctransaction.Request, trace string) (dict.Dict, error) {
	ch.Log.Debugf("runBatch ('%s')", trace)

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	ch.validateBatch(batch)

	// solidify arguments
	for _, req := range batch {
		if ok, err := req.SolidifyArgs(ch.Env.registry); err != nil || !ok {
			return nil, fmt.Errorf("solo inconsistency: failed to solidify request args")
		}
	}

	timestamp := ch.Env.LogicalTime().Add(time.Duration(len(batch)) * time.Nanosecond)
	task := &vm.VMTask{
		Processors:         ch.proc,
		ChainInput:         ch.GetChainOutput(),
		Requests:           batch,
		Timestamp:          timestamp,
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
	runvm.MustRunComputationsAsync(task)
	require.NoError(ch.Env.T, err)

	wg.Wait()
	inputs := sctransaction.OutputsFromRequests(task.Requests...)
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

	ch.State = newState

	ch.Log.Infof("state transition #%d --> #%d. Requests in the block: %d. Outputs: %d",
		ch.State.BlockIndex()-1, ch.State.BlockIndex(), len(block.RequestIDs()), len(stateTx.Essence().Outputs()))
	ch.Log.Debugf("Batch processed: %s", batchShortStr(block.RequestIDs()))

	ch.Env.EnqueueRequests(stateTx)
	ch.Env.ClockStep()
}

func batchShortStr(reqIds []*sctransaction.Request) string {
	ret := make([]string, len(reqIds))
	for i, r := range reqIds {
		ret[i] = r.Short()
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ","))
}
