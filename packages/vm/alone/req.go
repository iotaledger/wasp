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
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
	"strings"
	"sync"
	"time"
)

type callParams struct {
	target     coretypes.Hname
	entryPoint coretypes.Hname
	transfer   coretypes.ColoredBalances
	params     dict.Dict
}

func NewCall(target, ep string, params ...interface{}) *callParams {
	ret := &callParams{
		target:     coretypes.Hn(target),
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

func (e *AloneEnvironment) runBatch(batch []sctransaction.RequestRef, trace string) (dict.Dict, error) {
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
		Timestamp:    time.Now().UnixNano() + 1,
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
	prevBlockIndex := e.StateTx.MustState().BlockIndex()

	task.ResultTransaction.Sign(e.ChainSigScheme)
	err = e.UtxoDB.AddTransaction(task.ResultTransaction.Transaction)
	require.NoError(e.T, err)

	err = task.VirtualState.ApplyBlock(task.ResultBlock)
	require.NoError(e.T, err)

	err = task.VirtualState.CommitToDb(task.ResultBlock)
	require.NoError(e.T, err)

	e.StateTx = task.ResultTransaction
	e.State = task.VirtualState

	newBlockIndex := e.State.BlockIndex()

	e.Infof("state transition #%d --> #%d. Batch: %s. Posted requests: %d",
		prevBlockIndex, newBlockIndex, batchShortStr(batch), len(e.StateTx.Requests()))

	e.chPosted.Add(len(e.StateTx.Requests()))
	for i := range e.StateTx.Requests() {
		e.chInRequest <- sctransaction.RequestRef{
			Tx:    task.ResultTransaction,
			Index: uint16(i),
		}
	}
	return callRes, callErr
}

func batchShortStr(batch []sctransaction.RequestRef) string {
	ret := make([]string, len(batch))
	for i, r := range batch {
		ret[i] = r.RequestID().Short()
	}
	return fmt.Sprintf("[%s]", strings.Join(ret, ","))
}

func (e *AloneEnvironment) PostRequest(req *callParams, sigScheme signaturescheme.SignatureScheme) (dict.Dict, error) {
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

func (e *AloneEnvironment) CallView(req *callParams) (dict.Dict, error) {
	e.runVMMutex.Lock()
	defer e.runVMMutex.Unlock()

	vctx := viewcontext.New(e.ChainID, e.State.Variables(), e.Proc, e.Log)
	return vctx.CallView(req.target, req.entryPoint, req.params)
}
