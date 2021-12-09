// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"strings"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/stretchr/testify/require"
)

func (ch *Chain) runRequestsSync(reqs []iscp.RequestData, trace string) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	ch.mempool.ReceiveRequests(reqs...)
	ch.mempool.WaitInBufferEmpty()

	return ch.runRequestsNolock(reqs, trace)
}

func (ch *Chain) runRequestsNolock(reqs []iscp.RequestData, trace string) (dict.Dict, error) {
	ch.Log.Debugf("runRequestsNolock ('%s')", trace)

	anchorOutput, anchorOutputID := ch.GetAnchorOutput()
	task := &vm.VMTask{
		Processors:         ch.proc,
		AnchorOutput:       anchorOutput,
		AnchorOutputID:     *anchorOutputID,
		Requests:           reqs,
		TimeAssumption:     ch.Env.GlobalTime(),
		VirtualStateAccess: ch.State.Copy(),
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

	var essence *iotago.TransactionEssence

	if task.RotationAddress == nil {
		essence = task.ResultTransactionEssence
	} else {
		panic("not implemented")
		//essence, err = rotate.MakeRotateStateControllerTransaction(
		//	task.RotationAddress,
		//	task.AnchorOutput,
		//	task.Timestamp.Add(2*time.Nanosecond),
		//	identity.ID{},
		//	identity.ID{},
		//)
		//require.NoError(ch.Env.T, err)
	}

	txb := iotago.NewTransactionBuilder()
	for _, input := range essence.Inputs {
		txb.AddInput(&iotago.ToBeSignedUTXOInput{
			Address: ch.StateControllerAddress,
			Input:   input.(*iotago.UTXOInput),
		})
	}
	for _, out := range essence.Outputs {
		txb.AddOutput(out)
	}

	tx, err := txb.Build(
		parameters.DeSerializationParameters(),
		ch.StateControllerKeyPair.AsAddressSigner(),
	)
	require.NoError(ch.Env.T, err)

	err = ch.Env.AddToLedger(tx)
	require.NoError(ch.Env.T, err)

	// TODO: call settleStateTransition
	/*
		stateOutput, err := utxoutil.GetSingleChainedAliasOutput(tx)
		require.NoError(ch.Env.T, err)

		if task.RotationAddress == nil {
			// normal state transition
			ch.State = task.VirtualStateAccess
			ch.settleStateTransition(tx, stateOutput, iscp.TakeRequestIDs(reqs[0:task.ProcessedRequestsCount]...))
		} else {
			ch.Log.Infof("ROTATED STATE CONTROLLER to %s", stateOutput.GetStateAddress().Base58())
		}
	*/

	return callRes, callErr
}

func (ch *Chain) settleStateTransition(stateTx *iotago.Transaction, stateOutput *iotago.AliasOutput, reqids []iscp.RequestID) {
	panic("TODO implement")
	// stateOutput, err := utxoutil.GetSingleChainedAliasOutput(stateTx)
	// require.NoError(ch.Env.T, err)

	// // saving block just to check consistency. Otherwise, saved blocks are not used in Solo
	// block, err := ch.State.ExtractBlock()
	// require.NoError(ch.Env.T, err)
	// require.NotNil(ch.Env.T, block)
	// block.SetApprovingOutputID(stateOutput.ID())

	// err = ch.State.Commit(block)
	// require.NoError(ch.Env.T, err)

	// blockBack, err := state.LoadBlock(ch.Env.dbmanager.GetKVStore(ch.ChainID), ch.State.BlockIndex())
	// require.NoError(ch.Env.T, err)
	// require.True(ch.Env.T, bytes.Equal(block.Bytes(), blockBack.Bytes()))
	// require.EqualValues(ch.Env.T, stateOutput.ID(), blockBack.ApprovingOutputID())

	// chain.PublishStateTransition(ch.ChainID, stateOutput, len(reqids))
	// chain.PublishRequestsSettled(ch.ChainID, stateOutput.GetStateIndex(), reqids)

	// ch.Log.Infof("state transition --> #%d. Requests in the block: %d. Outputs: %d",
	// 	ch.State.BlockIndex(), len(reqids), len(stateTx.Essence().Outputs()))
	// ch.Log.Debugf("Batch processed: %s", batchShortStr(reqids))

	// ch.mempool.RemoveRequests(reqids...)

	// go ch.Env.EnqueueRequests(stateTx)
	// ch.Env.ClockStep()
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
