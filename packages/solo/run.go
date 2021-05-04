// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
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
		VirtualState:       ch.State,
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
	ch.settleStateTransition(tx)

	return callRes, callErr
}

func (ch *Chain) settleStateTransition(stateTx *ledgerstate.Transaction) {
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

	blockBack, err := state.LoadBlock(ch.Env.dbProvider, &ch.ChainID, ch.State.BlockIndex())
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, bytes.Equal(block.Bytes(), blockBack.Bytes()))
	require.EqualValues(ch.Env.T, stateOutput.ID(), blockBack.ApprovingOutputID())

	reqIDs := chain.PublishStateTransition(ch.State, stateOutput)

	ch.Log.Infof("state transition --> #%d. Requests in the block: %d. Outputs: %d",
		ch.State.BlockIndex(), len(reqIDs), len(stateTx.Essence().Outputs()))
	ch.Log.Debugf("Batch processed: %s", batchShortStr(reqIDs))

	ch.mempool.RemoveRequests(reqIDs...)

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
