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
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/rotate"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/stretchr/testify/require"
)

func (ch *Chain) runRequestsSync(reqs []iscp.Request, trace string) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	ch.mempool.ReceiveRequests(reqs...)
	ch.mempool.WaitInBufferEmpty()

	return ch.runRequestsNolock(reqs, trace)
}

func (ch *Chain) runRequestsNolock(reqs []iscp.Request, trace string) (dict.Dict, error) {
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
		VirtualState:       ch.State.Clone(),
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
	}

	ch.Env.vmRunner.Run(task)

	ch.Env.AdvanceClockBy(time.Duration(len(task.Requests)+1) * time.Nanosecond)

	var essence *ledgerstate.TransactionEssence

	if task.RotationAddress == nil {
		essence = task.ResultTransactionEssence
	} else {
		essence, err = rotate.MakeRotateStateControllerTransaction(
			task.RotationAddress,
			task.ChainInput,
			task.Timestamp.Add(2*time.Nanosecond),
			identity.ID{},
			identity.ID{},
		)
		require.NoError(ch.Env.T, err)
	}

	inputs, err := ch.Env.utxoDB.CollectUnspentOutputsFromInputs(essence)
	require.NoError(ch.Env.T, err)
	unlockBlocks, err := utxoutil.UnlockInputsWithED25519KeyPairs(inputs, essence, ch.StateControllerKeyPair)
	require.NoError(ch.Env.T, err)

	tx := ledgerstate.NewTransaction(essence, unlockBlocks)
	err = ch.Env.AddToLedger(tx)
	require.NoError(ch.Env.T, err)

	stateOutput, err := utxoutil.GetSingleChainedAliasOutput(tx)
	require.NoError(ch.Env.T, err)

	if task.RotationAddress == nil {
		// normal state transition
		ch.State = task.VirtualState
		ch.settleStateTransition(tx, stateOutput, iscp.TakeRequestIDs(reqs...))
	} else {
		ch.Log.Infof("ROTATED STATE CONTROLLER to %s", stateOutput.GetStateAddress().Base58())
	}

	return callRes, callErr
}

//nolint // TODO check this function, the `stateOutput` param is unused, and its re-assigned on the first line
func (ch *Chain) settleStateTransition(stateTx *ledgerstate.Transaction, stateOutput *ledgerstate.AliasOutput, reqids []iscp.RequestID) {
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

	chain.PublishStateTransition(&ch.ChainID, stateOutput, len(reqids))
	chain.PublishRequestsSettled(&ch.ChainID, stateOutput.GetStateIndex(), reqids)

	ch.Log.Infof("state transition --> #%d. Requests in the block: %d. Outputs: %d",
		ch.State.BlockIndex(), len(reqids), len(stateTx.Essence().Outputs()))
	ch.Log.Debugf("Batch processed: %s", batchShortStr(reqids))

	ch.mempool.RemoveRequests(reqids...)

	go ch.Env.EnqueueRequests(stateTx)
	ch.Env.ClockStep()
}

func batchShortStr(reqIds []iscp.RequestID) string {
	ret := make([]string, len(reqIds))
	for i, r := range reqIds {
		ret[i] = r.Short()
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ","))
}

func (ch *Chain) logRequestLastBlock() {
	recs := ch.GetRequestReceiptsForBlock(ch.GetLatestBlockInfo().BlockIndex)
	for _, rec := range recs {
		ch.Log.Infof("REQ: '%s'", rec.Short())
	}
}
