// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/identity"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"

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

	ch.mempool.ReceiveRequests(reqs...)
	ch.mempool.WaitInBufferEmpty()

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
	isRotateRequest := callErr == nil && len(task.Requests) == 1 && coreutil.IsRotateCommitteeRequest(task.Requests[0])

	ts := task.ResultTransactionEssence.Timestamp()

	if !isRotateRequest {
		essence = task.ResultTransactionEssence
	} else {
		req := task.Requests[0]
		par, _ := req.Params()
		deco := kvdecoder.New(par, ch.Log)
		nextAddr := deco.MustGetAddress(coreutil.ParamStateControllerAddress)

		essence, err = coreutil.MakeRotateStateControllerTransaction(
			nextAddr,
			task.ChainInput,
			task.Requests[0].Output(),
			ts,
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

	if !isRotateRequest {
		ch.State = task.VirtualState
		ch.settleStateTransition(tx, stateOutput, coretypes.TakeRequestIDs(reqs...))
	} else {
		ch.Log.Infof("ROTATED STATE CONTROLLER to %s from prev %s",
			stateOutput.GetStateAddress().Base58(), ch.StateControllerAddress.Base58())
		ch.mempool.RemoveRequests(task.Requests[0].ID())
	}

	return callRes, callErr
}

func (ch *Chain) settleStateTransition(stateTx *ledgerstate.Transaction, stateOutput *ledgerstate.AliasOutput, reqids []coretypes.RequestID) {
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
