// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
)

type callParams struct {
	targetName string
	target     coretypes.Hname
	epName     string
	entryPoint coretypes.Hname
	transfer   coretypes.ColoredBalances
	params     dict.Dict
}

func NewCall(sc, fun string, params ...interface{}) *callParams {
	ret := &callParams{
		targetName: sc,
		target:     coretypes.Hn(sc),
		epName:     fun,
		entryPoint: coretypes.Hn(fun),
	}
	ret.withParams(params...)
	return ret
}

func NewCallFromDict(sc, fun string, params dict.Dict) *callParams {
	par := make([]interface{}, 0, 2*len(params))
	params.MustIterate("", func(key kv.Key, value []byte) bool {
		par = append(par, string(key))
		par = append(par, value)
		return true
	})
	return NewCall(sc, fun, par...)
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

func (ch *Chain) runBatch(batch []sctransaction.RequestRef, trace string) (dict.Dict, error) {
	ch.Log.Debugf("runBatch ('%s'): %s", trace, batchShortStr(batch))
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	task := &vm.VMTask{
		Processors:         ch.Proc,
		ChainID:            ch.ChainID,
		Color:              ch.ChainColor,
		Entropy:            *hashing.RandomHash(nil),
		ValidatorFeeTarget: ch.ValidatorFeeTarget,
		Balances:           waspconn.OutputsToBalances(ch.Glb.utxoDB.GetAddressOutputs(ch.ChainAddress)),
		Requests:           batch,
		Timestamp:          ch.Glb.LogicalTime().UnixNano(),
		VirtualState:       ch.State.Clone(),
		Log:                ch.Log,
	}
	var err error
	var wg sync.WaitGroup
	var callRes dict.Dict
	var callErr error
	task.OnFinish = func(callResult dict.Dict, callError error, err error) {
		require.NoError(ch.Glb.T, err)
		callRes = callResult
		callErr = callError
		wg.Done()
	}

	wg.Add(1)
	err = runvm.RunComputationsAsync(task)
	require.NoError(ch.Glb.T, err)

	wg.Wait()
	task.ResultTransaction.Sign(ch.ChainSigScheme)
	prevBlockIndex := ch.StateTx.MustState().BlockIndex()

	ch.settleStateTransition(task.VirtualState, task.ResultBlock, task.ResultTransaction)

	ch.Infof("state transition #%d --> #%d. Batch: %s. Posted requests: %d",
		prevBlockIndex, ch.State.BlockIndex(), batchShortStr(batch), len(ch.StateTx.Requests()))
	return callRes, callErr
}

func (ch *Chain) settleStateTransition(newState state.VirtualState, block state.Block, stateTx *sctransaction.Transaction) {
	err := ch.Glb.utxoDB.AddTransaction(stateTx.Transaction)
	require.NoError(ch.Glb.T, err)

	err = newState.ApplyBlock(block)
	require.NoError(ch.Glb.T, err)

	err = newState.CommitToDb(block)
	require.NoError(ch.Glb.T, err)

	ch.StateTx = stateTx
	ch.State = newState
	ch.Glb.ClockStep()

	// dispatch requests among chains
	ch.Glb.glbMutex.Lock()
	defer ch.Glb.glbMutex.Unlock()

	reqRefByChain := make(map[coretypes.ChainID][]sctransaction.RequestRef)
	for i, rsect := range ch.StateTx.Requests() {
		chid := rsect.Target().ChainID()
		_, ok := reqRefByChain[chid]
		if !ok {
			reqRefByChain[chid] = make([]sctransaction.RequestRef, 0)
		}
		reqRefByChain[chid] = append(reqRefByChain[chid], sctransaction.RequestRef{
			Tx:    stateTx,
			Index: uint16(i),
		})
	}
	for chid, reqs := range reqRefByChain {
		chain, ok := ch.Glb.chains[chid]
		if !ok {
			ch.Infof("dispatching requests. Unknown chain: %s", chid.String())
			continue
		}
		chain.chPosted.Add(len(reqs))
		for _, reqRef := range reqs {
			chain.chInRequest <- reqRef
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

func (ch *Chain) PostRequest(req *callParams, sigScheme signaturescheme.SignatureScheme) (dict.Dict, error) {
	if sigScheme == nil {
		sigScheme = ch.OriginatorSigScheme
	}
	allOuts := ch.Glb.utxoDB.GetAddressOutputs(sigScheme.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(ch.Glb.T, err)

	reqSect := sctransaction.NewRequestSectionByWallet(coretypes.NewContractID(ch.ChainID, req.target), req.entryPoint).
		WithTransfer(req.transfer).
		WithArgs(req.params)
	err = txb.AddRequestSection(reqSect)
	require.NoError(ch.Glb.T, err)

	tx, err := txb.Build(false)
	require.NoError(ch.Glb.T, err)

	tx.Sign(sigScheme)
	err = ch.Glb.utxoDB.AddTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}
	reqID := coretypes.NewRequestID(tx.ID(), 0)
	ch.Log.Infof("PostRequest: %s::%s -- %s", req.targetName, req.epName, reqID.String())
	return ch.runBatch([]sctransaction.RequestRef{{Tx: tx, Index: 0}}, "post")
}

func (ch *Chain) CallViewFull(req *callParams) (dict.Dict, error) {
	ch.Log.Infof("CallViewFull: %s::%s", req.targetName, req.epName)
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.State.Variables(), ch.State.Timestamp(), ch.Proc, ch.Log)
	return vctx.CallView(req.target, req.entryPoint, req.params)
}

func (ch *Chain) CallView(sc string, fun string, params ...interface{}) (dict.Dict, error) {
	return ch.CallViewFull(NewCall(sc, fun, params...))
}

func (ch *Chain) WaitEmptyBacklog(maxWait ...time.Duration) {
	maxDurationSet := len(maxWait) > 0
	var deadline time.Time
	if maxDurationSet {
		deadline = time.Now().Add(maxWait[0])
	}
	counter := 0
	for {
		if counter%40 == 0 {
			ch.Log.Infof("backlog length = %d", ch.backlogLen())
		}
		counter++
		if ch.backlogLen() > 0 {
			time.Sleep(50 * time.Millisecond)
			if maxDurationSet && deadline.Before(time.Now()) {
				ch.Log.Warnf("exit due to timeout of max wait for %v", maxWait[0])
			}
		} else {
			time.Sleep(10 * time.Millisecond)
			if ch.backlogLen() == 0 {
				break
			}
		}
	}
}
