package alone

import (
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
	"sync"
	"time"
)

type callParams struct {
	target     coretypes.Hname
	entryPoint coretypes.Hname
	transfer   coretypes.ColoredBalances
	params     dict.Dict
}

func NewCall(target, ep string) *callParams {
	return &callParams{
		target:     coretypes.Hn(target),
		entryPoint: coretypes.Hn(ep),
	}
}

func (r *callParams) WithTransfer(transfer map[balance.Color]int64) *callParams {
	r.transfer = cbalances.NewFromMap(transfer)
	return r
}
func (r *callParams) WithParams(params ...interface{}) *callParams {
	if len(params) == 0 {
		return r
	}
	if len(params)%2 != 0 {
		panic("WithParams: len(params) % 2 != 0")
	}
	par := make(map[string]interface{})
	for i := 0; i < len(params)/2; i++ {
		key, ok := params[2*i].(string)
		if !ok {
			panic("WithParams: string expected")
		}
		par[key] = params[2*i+1]
	}
	r.params = codec.MakeDict(par)
	return r
}

func (e *aloneEnvironment) runRequest(reqTx *sctransaction.Transaction) (dict.Dict, error) {
	err := e.UtxoDB.AddTransaction(reqTx.Transaction)
	require.NoError(e.T, err)

	reqRef := sctransaction.RequestRef{Tx: reqTx}
	task := &vm.VMTask{
		Processors:   e.Proc,
		ChainID:      e.ChainID,
		Color:        e.ChainColor,
		Entropy:      *hashing.RandomHash(nil),
		Balances:     waspconn.OutputsToBalances(e.UtxoDB.GetAddressOutputs(e.ChainAddress)),
		Requests:     []sctransaction.RequestRef{reqRef},
		Timestamp:    time.Now().UnixNano() + 1,
		VirtualState: e.State.Clone(),
		Log:          e.Log,
	}

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

	task.ResultTransaction.Sign(e.ChainSigscheme)
	err = e.UtxoDB.AddTransaction(task.ResultTransaction.Transaction)
	require.NoError(e.T, err)

	err = task.VirtualState.ApplyBlock(task.ResultBlock)
	require.NoError(e.T, err)

	err = task.VirtualState.CommitToDb(task.ResultBlock)
	require.NoError(e.T, err)

	e.StateTx = task.ResultTransaction
	e.State = task.VirtualState

	newBlockIndex := e.State.BlockIndex()
	e.Infof("state transition #%d --> #%d. Req: %s", prevBlockIndex, newBlockIndex, reqRef.RequestID().String())

	return callRes, callErr
}

func (e *aloneEnvironment) PostRequest(req *callParams, sigScheme signaturescheme.SignatureScheme) (dict.Dict, error) {
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
	return e.runRequest(tx)
}

func (e *aloneEnvironment) CallView(req *callParams) (dict.Dict, error) {
	vctx := viewcontext.New(e.ChainID, e.State.Variables(), e.Proc, e.Log)
	return vctx.CallView(req.target, req.entryPoint, req.params)
}
