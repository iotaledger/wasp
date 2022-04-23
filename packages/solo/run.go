// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"
	"fmt"
	"strings"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func (ch *Chain) runRequestsSync(reqs []iscp.Request, trace string) (results []*vm.RequestResult) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	ch.mempool.ReceiveRequests(reqs...)
	ch.mempool.WaitInBufferEmpty()

	return ch.runRequestsNolock(reqs, trace)
}

func (ch *Chain) estimateGas(req iscp.Request) (result *vm.RequestResult) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	task := ch.runTaskNoLock([]iscp.Request{req}, true)
	require.Len(ch.Env.T, task.Results, 1, "cannot estimate gas: request was skipped")
	return task.Results[0]
}

func (ch *Chain) runTaskNoLock(reqs []iscp.Request, estimateGas bool) *vm.VMTask {
	anchorOutput := ch.GetAnchorOutput()
	task := &vm.VMTask{
		Processors:         ch.proc,
		AnchorOutput:       anchorOutput.GetAliasOutput(),
		AnchorOutputID:     anchorOutput.OutputID(),
		Requests:           reqs,
		TimeAssumption:     ch.Env.GlobalTime(),
		VirtualStateAccess: ch.State.Copy(),
		Entropy:            hashing.RandomHash(nil),
		ValidatorFeeTarget: ch.ValidatorFeeTarget,
		Log:                ch.Log().Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
		L1Params:           ch.Env.utxoDB.L1Params(),
		// state baseline is always valid in Solo
		SolidStateBaseline:   ch.GlobalSync.GetSolidIndexBaseline(),
		EnableGasBurnLogging: true,
		EstimateGasMode:      estimateGas,
	}

	ch.Env.vmRunner.Run(task)
	require.NoError(ch.Env.T, task.VMError)

	return task
}

func (ch *Chain) runRequestsNolock(reqs []iscp.Request, trace string) (results []*vm.RequestResult) {
	ch.Log().Debugf("runRequestsNolock ('%s')", trace)

	task := ch.runTaskNoLock(reqs, false)

	var essence *iotago.TransactionEssence
	var inputsCommitment []byte
	if task.RotationAddress == nil {
		essence = task.ResultTransactionEssence
		inputsCommitment = task.ResultInputsCommitment
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
	sigs, err := essence.Sign(
		inputsCommitment,
		ch.StateControllerKeyPair.GetPrivateKey().AddressKeys(ch.StateControllerAddress),
	)
	require.NoError(ch.Env.T, err)

	tx := transaction.MakeAnchorTransaction(essence, sigs[0])

	err = ch.Env.AddToLedger(tx)
	require.NoError(ch.Env.T, err)

	anchor, _, err := transaction.GetAnchorFromTransaction(tx)

	if task.RotationAddress == nil {
		// normal state transition
		ch.State = task.VirtualStateAccess
		ch.settleStateTransition(tx, task.GetProcessedRequestIDs())
	} else {
		require.NoError(ch.Env.T, err)

		ch.Log().Infof("ROTATED STATE CONTROLLER to %s", anchor.StateController)
	}

	rootC := ch.GetRootCommitment()
	l1C := ch.GetL1Commitment()
	require.True(ch.Env.T, trie.EqualCommitments(rootC, l1C.StateCommitment))

	return task.Results
}

func (ch *Chain) settleStateTransition(stateTx *iotago.Transaction, reqids []iscp.RequestID) {
	anchor, stateOutput, err := transaction.GetAnchorFromTransaction(stateTx)
	require.NoError(ch.Env.T, err)

	// saving block just to check consistency. Otherwise, saved blocks are not used in Solo
	block, err := ch.State.ExtractBlock()
	require.NoError(ch.Env.T, err)
	require.NotNil(ch.Env.T, block)
	block.SetApprovingOutputID(anchor.OutputID.UTXOInput())

	err = ch.State.Save(block)
	require.NoError(ch.Env.T, err)

	blockBack, err := state.LoadBlock(ch.Env.dbmanager.GetKVStore(ch.ChainID), ch.State.BlockIndex())
	require.NoError(ch.Env.T, err)
	require.True(ch.Env.T, bytes.Equal(block.Bytes(), blockBack.Bytes()))
	require.EqualValues(ch.Env.T, anchor.OutputID, blockBack.ApprovingOutputID().ID())

	chain.PublishStateTransition(ch.ChainID, iscp.NewAliasOutputWithID(stateOutput, anchor.OutputID.UTXOInput()), len(reqids))
	chain.PublishRequestsSettled(ch.ChainID, anchor.StateIndex, reqids)

	ch.Log().Infof("state transition --> #%d. Requests in the block: %d. Outputs: %d",
		ch.State.BlockIndex(), len(reqids), len(stateTx.Essence.Outputs))
	ch.Log().Debugf("Batch processed: %s", batchShortStr(reqids))

	ch.mempool.RemoveRequests(reqids...)

	go ch.Env.EnqueueRequests(stateTx)
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
		ch.Log().Infof("REQ: '%s'", rec.Short())
	}
}
