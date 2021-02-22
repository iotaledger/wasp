// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
)

type CallParams struct {
	targetName string
	target     coretypes.Hname
	epName     string
	entryPoint coretypes.Hname
	transfer   coretypes.ColoredBalances
	mint       map[address.Address]int64
	args       requestargs.RequestArgs
}

func NewCallParamsFromDic(scName, funName string, par dict.Dict) *CallParams {
	ret := &CallParams{
		targetName: scName,
		target:     coretypes.Hn(scName),
		epName:     funName,
		entryPoint: coretypes.Hn(funName),
	}
	ret.args = requestargs.New(nil)
	for k, v := range par {
		ret.args.AddEncodeSimple(k, v)
	}
	return ret
}

// NewCallParams creates structure which wraps in one object call parameters, used in PostRequestSync and callViewFull
// calls:
//  - 'scName' is a a name of the target smart contract
//  - 'funName' is a name of the target entry point (the function) of he smart contract program
//  - 'params' is a sequence of pairs 'paramName', 'paramValue' which constitute call parameters
//     The 'paramName' must be a string and 'paramValue' must different types (encoded based on type)
// With the WithTransfers the CallParams structure may be complemented with attached colored
// tokens sent together with the request
func NewCallParams(scName, funName string, params ...interface{}) *CallParams {
	return NewCallParamsFromDic(scName, funName, codec.MakeDict(toMap(params...)))
}

func NewCallParamsOptimized(scName, funName string, optSize int, params ...interface{}) (*CallParams, map[kv.Key][]byte) {
	if optSize <= 32 {
		optSize = 32
	}
	ret := &CallParams{
		targetName: scName,
		target:     coretypes.Hn(scName),
		epName:     funName,
		entryPoint: coretypes.Hn(funName),
	}
	d := codec.MakeDict(toMap(params...))
	var retOptimized map[kv.Key][]byte
	ret.args, retOptimized = requestargs.NewOptimizedRequestArgs(d)
	return ret, retOptimized
}

// WithTransfer is a shorthand for the most often used case where only
// a single color is transferred by WithTransfers
func (r *CallParams) WithTransfer(color balance.Color, amount int64) *CallParams {
	return r.WithTransfers(map[balance.Color]int64{color: amount})
}

// WithTransfers complement CallParams structure with the colored balances of tokens
// in the form of a collection of pairs 'color': 'balance'
func (r *CallParams) WithTransfers(transfer map[balance.Color]int64) *CallParams {
	r.transfer = cbalances.NewFromMap(transfer)
	return r
}

// WithMinting adds minting part to the request transaction
// in the map <address>: <amount> all addresses should be different from the
// address of the target chain and amounts must be positive
func (r *CallParams) WithMinting(mint map[address.Address]int64) *CallParams {
	r.mint = make(map[address.Address]int64)
	for addr, amount := range mint {
		r.mint[addr] = amount
	}
	return r
}

// makes map without hashing
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

// RequestFromParamsToLedger creates transaction with one request based on parameters and sigScheme
// Then it adds it to the ledger, atomically.
// Locking on the mutex is needed to prevent mess when several goroutines work on he same address
func (ch *Chain) RequestFromParamsToLedger(req *CallParams, sigScheme signaturescheme.SignatureScheme) *sctransaction.Transaction {
	ch.Env.ledgerMutex.Lock()
	defer ch.Env.ledgerMutex.Unlock()

	if sigScheme == nil {
		sigScheme = ch.OriginatorSigScheme
	}
	allOuts := ch.Env.utxoDB.GetAddressOutputs(sigScheme.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(ch.Env.T, err)

	reqSect := sctransaction.NewRequestSectionByWallet(coretypes.NewContractID(ch.ChainID, req.target), req.entryPoint).
		WithTransfer(req.transfer).
		WithArgs(req.args)

	err = txb.AddRequestSection(reqSect)
	require.NoError(ch.Env.T, err)

	txb.AddMinting(req.mint)

	tx, err := txb.Build(false)
	require.NoError(ch.Env.T, err)

	tx.Sign(sigScheme)

	_, err = tx.Properties()
	require.NoError(ch.Env.T, err)

	err = ch.Env.AddToLedger(tx)
	require.NoError(ch.Env.T, err)
	return tx
}

// PostRequestSync posts a request synchronously  sent by the test program to the smart contract on the same or another chain:
//  - creates a request transaction with the request block on it. The sigScheme is used to
//    sign the inputs of the transaction or OriginatorSigScheme is used if parameter is nil
//  - adds request transaction to UTXODB
//  - runs the request in the VM. It results in new updated virtual state and a new transaction
//    which anchors the state.
//  - adds the resulting transaction to UTXODB
//  - posts requests, contained in the resulting transaction to backlog queues of respective chains
//  - returns the result of the call to the smart contract's entry point
// Note that in real network of Wasp nodes (the committee) posting the transaction is completely
// asynchronous, i.e. result of the call is not available to the originator of the post.
//
// Unlike the real Wasp environment, the 'solo' environment makes PostRequestSync a synchronous call.
// It makes it possible step-by-step debug of the smart contract logic.
// The call should be used only from the main thread (goroutine)
func (ch *Chain) PostRequestSync(req *CallParams, sigScheme signaturescheme.SignatureScheme) (dict.Dict, error) {
	_, ret, err := ch.PostRequestSyncTx(req, sigScheme)
	return ret, err
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, sigScheme signaturescheme.SignatureScheme) (*sctransaction.Transaction, dict.Dict, error) {
	tx := ch.RequestFromParamsToLedger(req, sigScheme)

	reqID := coretypes.NewRequestID(tx.ID(), 0)
	ch.Log.Infof("PostRequestSync: %s::%s -- %s", req.targetName, req.epName, reqID.String())

	r := vm.RequestRefWithFreeTokens{}
	r.Tx = tx
	ch.reqCounter.Add(1)
	ret, err := ch.runBatch([]vm.RequestRefWithFreeTokens{r}, "post")
	if err != nil {
		return nil, nil, err
	}
	return tx, ret, nil
}

// callViewFull calls the view entry point of the smart contract
// with params wrapped into the CallParams object. The transfer part, fs any, is ignored
func (ch *Chain) callViewFull(req *CallParams) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.State.Variables(), ch.State.Timestamp(), ch.proc, ch.Log)
	a, ok, err := req.args.SolidifyRequestArguments(ch.Env.registry)
	if err != nil || !ok {
		return nil, fmt.Errorf("solo.internal error: can't solidify args")
	}
	return vctx.CallView(req.target, req.entryPoint, a)
}

// CallView calls the view entry point of the smart contract.
// The call params should be in pairs ('paramName', 'paramValue') where 'paramName' is a string
// and 'paramValue' must be of type accepted by the 'codec' package
func (ch *Chain) CallView(scName string, funName string, params ...interface{}) (dict.Dict, error) {
	ch.Log.Infof("callView: %s::%s", scName, funName)

	p := codec.MakeDict(toMap(params...))

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.State.Variables(), ch.State.Timestamp(), ch.proc, ch.Log)
	return vctx.CallView(coretypes.Hn(scName), coretypes.Hn(funName), p)
}

// WaitForEmptyBacklog waits until the backlog queue of the chain becomes empty.
// It is useful when smart contract(s) in the test are posting asynchronous requests
// between chains.
//
// The call is needed in order to prevent finishing the test before all
// asynchronous request between chains are processed.
// Otherwise waiting is not necessary because all PostRequestSync calls by the test itself
// are synchronous and are processed immediately
func (ch *Chain) WaitForEmptyBacklog(maxWait ...time.Duration) {
	maxw := 5 * time.Second
	var deadline time.Time
	if len(maxWait) > 0 {
		maxw = maxWait[0]
	}
	deadline = time.Now().Add(maxw)
	counter := 0
	for {
		if counter%40 == 0 {
			ch.Log.Infof("backlog length = %d", ch.backlogLen())
		}
		counter++
		time.Sleep(200 * time.Millisecond)
		if ch.backlogLen() > 0 {
			if time.Now().After(deadline) {
				ch.Log.Warnf("exit due to timeout of max wait for %v", maxw)
				return
			}
		} else {
			emptyCounter := 0
			for i := 0; i < 5; i++ {
				time.Sleep(100 * time.Millisecond)
				if ch.backlogLen() != 0 {
					break
				}
				emptyCounter++
			}
			if emptyCounter >= 5 {
				return
			}
		}
	}
}
