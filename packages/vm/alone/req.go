// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package alone

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"time"
)

type callParams struct {
	targetName string
	target     coretypes.Hname
	epName     string
	entryPoint coretypes.Hname
	transfer   coretypes.ColoredBalances
	params     dict.Dict
}

func NewCall(target, ep string, params ...interface{}) *callParams {
	ret := &callParams{
		targetName: target,
		target:     coretypes.Hn(target),
		epName:     ep,
		entryPoint: coretypes.Hn(ep),
	}
	ret.withParams(params...)
	return ret
}

func (r *callParams) WithTransfer(transfer map[balance.Color]int64) *callParams {
	r.transfer = cbalances.NewFromMap(transfer)
	return r
}

func toMap(params ...interface{}) map[string]interface{} {
	par := make(map[string]interface{})
	if len(params) == 0 {
		return par
	}
	if len(params)%2 != 0 {
		panic("WithParams: len(params) % 2 != 0")
	}
	for i := 0; i < len(params)/2; i++ {
		key, ok := params[2*i].(string)
		if !ok {
			panic("WithParams: string expected")
		}
		par[key] = params[2*i+1]
	}
	return par
}

func (r *callParams) withParams(params ...interface{}) *callParams {
	r.params = codec.MakeDict(toMap(params...))
	return r
}

func (e *Env) runBatch(batch []sctransaction.RequestRef, trace string) (dict.Dict, error) {
	e.Log.Debugf("runBatch ('%s'): %s", trace, batchShortStr(batch))
	e.runVMMutex.Lock()
	defer e.runVMMutex.Unlock()

	task := &vm.VMTask{
		Processors:   e.Proc,
		ChainID:      e.ChainID,
		Color:        e.ChainColor,
		Entropy:      *hashing.RandomHash(nil),
		Balances:     waspconn.OutputsToBalances(e.UtxoDB.GetAddressOutputs(e.ChainAddress)),
		Requests:     batch,
		Timestamp:    e.timestamp.UnixNano(),
		VirtualState: e.State.Clone(),
		Log:          e.Log,
	}
	var err error
	var wg sync.WaitGroup
	var callRes dict.Dict
	var callErr error
	task.OnFinish = func(callResult dict.Dict, callError error, err error) {
		require.NoError(e.T, err)
		callRes = callResult
		callErr = callError
		wg.Done()
	}

	wg.Add(1)
	err = runvm.RunComputationsAsync(task)
	require.NoError(e.T, err)

	wg.Wait()
	task.ResultTransaction.Sign(e.ChainSigScheme)
	prevBlockIndex := e.StateTx.MustState().BlockIndex()

	e.settleStateTransition(task.VirtualState, task.ResultBlock, task.ResultTransaction)

	e.Infof("state transition #%d --> #%d. Batch: %s. Posted requests: %d",
		prevBlockIndex, e.State.BlockIndex(), batchShortStr(batch), len(e.StateTx.Requests()))
	return callRes, callErr
}

func (e *Env) settleStateTransition(newState state.VirtualState, block state.Block, stateTx *sctransaction.Transaction) {
	err := e.UtxoDB.AddTransaction(stateTx.Transaction)
	require.NoError(e.T, err)

	err = newState.ApplyBlock(block)
	require.NoError(e.T, err)

	err = newState.CommitToDb(block)
	require.NoError(e.T, err)

	e.StateTx = stateTx
	e.State = newState
	e.AdvanceClockBy(e.timeStep)

	e.chPosted.Add(len(e.StateTx.Requests()))
	for i := range e.StateTx.Requests() {
		e.chInRequest <- sctransaction.RequestRef{
			Tx:    stateTx,
			Index: uint16(i),
		}
	}
}

func batchShortStr(batch []sctransaction.RequestRef) string {
	ret := make([]string, len(batch))
	for i, r := range batch {
		ret[i] = r.RequestID().Short()
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ","))
}

func (e *Env) PostRequest(req *callParams, sigScheme signaturescheme.SignatureScheme) (dict.Dict, error) {
	if sigScheme == nil {
		sigScheme = e.OriginatorSigScheme
	}
	allOuts := e.UtxoDB.GetAddressOutputs(sigScheme.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(e.T, err)

	reqSect := sctransaction.NewRequestSectionByWallet(coretypes.NewContractID(e.ChainID, req.target), req.entryPoint).
		WithTransfer(req.transfer).
		WithArgs(req.params)
	err = txb.AddRequestSection(reqSect)
	require.NoError(e.T, err)

	tx, err := txb.Build(false)
	require.NoError(e.T, err)

	tx.Sign(sigScheme)
	err = e.UtxoDB.AddTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}
	return e.runBatch([]sctransaction.RequestRef{{Tx: tx, Index: 0}}, "post")
}

func (e *Env) CallViewFull(req *callParams) (dict.Dict, error) {
	e.runVMMutex.Lock()
	defer e.runVMMutex.Unlock()

	vctx := viewcontext.New(e.ChainID, e.State.Variables(), e.State.Timestamp(), e.Proc, e.Log)
	return vctx.CallView(req.target, req.entryPoint, req.params)
}

func (e *Env) CallView(fun string, ep string, params ...interface{}) (dict.Dict, error) {
	return e.CallViewFull(NewCall(fun, ep, params...))
}

func (e *Env) WaitEmptyBacklog(maxWait ...time.Duration) {
	maxDurationSet := len(maxWait) > 0
	var deadline time.Time
	if maxDurationSet {
		deadline = time.Now().Add(maxWait[0])
	}
	counter := 0
	for {
		if counter%40 == 0 {
			e.Log.Infof("backlog length = %d", e.backlogLen())
		}
		counter++
		if e.backlogLen() > 0 {
			time.Sleep(50 * time.Millisecond)
			if maxDurationSet && deadline.Before(time.Now()) {
				e.Log.Warnf("exit due to timeout of max wait for %v", maxWait[0])
			}
		} else {
			time.Sleep(10 * time.Millisecond)
			if e.backlogLen() == 0 {
				break
			}
		}
	}
}
