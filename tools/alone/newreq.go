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
	"github.com/stretchr/testify/require"
	"sync"
	"time"
)

type Request struct {
	sigScheme  signaturescheme.SignatureScheme
	target     coretypes.Hname
	entryPoint coretypes.Hname
	transfer   coretypes.ColoredBalances
	params     dict.Dict
}

func NewRequest(sigScheme signaturescheme.SignatureScheme, target, ep string) *Request {
	return &Request{
		sigScheme:  sigScheme,
		target:     coretypes.Hn(target),
		entryPoint: coretypes.Hn(ep),
	}
}

func (r *Request) WithTransfer(transfer map[balance.Color]int64) *Request {
	r.transfer = cbalances.NewFromMap(transfer)
	return r
}
func (r *Request) WithParams(params map[string]interface{}) *Request {
	r.params = codec.MakeDict(params)
	return r
}

func (e *Environment) runRequest(reqTx *sctransaction.Transaction) (dict.Dict, error) {
	err := Env.UtxoDB.AddTransaction(reqTx.Transaction)
	require.NoError(Env.T, err)

	task := &vm.VMTask{
		Processors:   Env.Proc,
		ChainID:      Env.ChainID,
		Color:        Env.ChainColor,
		Entropy:      *hashing.RandomHash(nil),
		Balances:     waspconn.OutputsToBalances(Env.UtxoDB.GetAddressOutputs(Env.ChainAddress)),
		Requests:     []sctransaction.RequestRef{{Tx: reqTx}},
		Timestamp:    time.Now().UnixNano() + 1,
		VirtualState: Env.State.Clone(),
		Log:          Env.Log,
	}

	var wg sync.WaitGroup
	var callRes dict.Dict
	var callErr error
	task.OnFinish = func(callResult dict.Dict, callError error, err error) {
		require.NoError(Env.T, err)
		callRes = callResult
		callErr = callError
		wg.Done()
	}

	wg.Add(1)
	err = runvm.RunComputationsAsync(task)
	require.NoError(Env.T, err)

	wg.Wait()
	prevBlockIndex := Env.StateTx.MustState().BlockIndex()

	task.ResultTransaction.Sign(Env.ChainSigscheme)
	err = Env.UtxoDB.AddTransaction(task.ResultTransaction.Transaction)
	require.NoError(Env.T, err)

	Env.StateTx = task.ResultTransaction
	Env.State = task.VirtualState
	newBlockIndex := Env.StateTx.MustState().BlockIndex()
	Env.Infof("state transition #%d --> #%d", prevBlockIndex, newBlockIndex)

	return callRes, callErr
}

func (e *Environment) PostRequest(req *Request) (dict.Dict, error) {
	allOuts := e.UtxoDB.GetAddressOutputs(req.sigScheme.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(e.T, err)

	reqSect := sctransaction.NewRequestSectionByWallet(coretypes.NewContractID(Env.ChainID, req.target), req.entryPoint).
		WithTransfer(req.transfer).
		WithArgs(req.params)
	err = txb.AddRequestSection(reqSect)
	require.NoError(e.T, err)

	tx, err := txb.Build(false)
	require.NoError(e.T, err)

	tx.Sign(req.sigScheme)
	return e.runRequest(tx)
}
