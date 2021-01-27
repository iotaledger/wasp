// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
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
	params     dict.Dict
}

// NewCall creates structure which wraps in one object call parameters, used in PostRequest and callViewFull
// calls:
//  - 'scName' is a a name of the target smart contract
//  - 'funName' is a name of the target entry point (the function) of he smart contract program
//  - 'params' is a sequence of pairs 'paramName', 'paramValue' which constitute call parameters
//     The 'paramName' must be a string and 'paramValue' must different types (encoded based on type)
// With the WithTransfers the CallParams structure may be complemented with attached colored
// tokens sent together with the request
func NewCall(scName, funName string, params ...interface{}) *CallParams {
	ret := &CallParams{
		targetName: scName,
		target:     coretypes.Hn(scName),
		epName:     funName,
		entryPoint: coretypes.Hn(funName),
	}
	ret.withParams(params...)
	return ret
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

func (r *CallParams) withParams(params ...interface{}) *CallParams {
	r.params = codec.MakeDict(toMap(params...))
	return r
}

// PostRequest posts a request sent by the test program to the smart contract on the same or another chain:
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
// Unlike the real Wasp environment, the 'solo' environment makes PostRequest a synchronous call.
// It makes it possible step-by-step debug of the smart contract logic.
func (ch *Chain) PostRequest(req *CallParams, sigScheme signaturescheme.SignatureScheme) (dict.Dict, error) {
	if sigScheme == nil {
		sigScheme = ch.OriginatorSigScheme
	}
	allOuts := ch.Env.utxoDB.GetAddressOutputs(sigScheme.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(ch.Env.T, err)

	reqSect := sctransaction.NewRequestSectionByWallet(coretypes.NewContractID(ch.ChainID, req.target), req.entryPoint).
		WithTransfer(req.transfer).
		AddArgs(req.params)
	err = txb.AddRequestSection(reqSect)
	require.NoError(ch.Env.T, err)

	tx, err := txb.Build(false)
	require.NoError(ch.Env.T, err)

	tx.Sign(sigScheme)
	err = ch.Env.utxoDB.AddTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}

	reqID := coretypes.NewRequestID(tx.ID(), 0)
	ch.Log.Infof("PostRequest: %s::%s -- %s", req.targetName, req.epName, reqID.String())

	r := vm.RequestRefWithFreeTokens{}
	r.Tx = tx
	return ch.runBatch([]vm.RequestRefWithFreeTokens{r}, "post")
}

// callViewFull calls the view entry point of the smart contract
// with params wrapped into the CallParams object. The transfer part, is any, is ignored
func (ch *Chain) callViewFull(req *CallParams) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.State.Variables(), ch.State.Timestamp(), ch.proc, ch.Log)
	return vctx.CallView(req.target, req.entryPoint, req.params)
}

// CallView calls the view entry point of the smart contract.
// The call params should be in pairs ('paramName', 'paramValue') where 'paramName' is a string
// and 'paramValue' must be of type accepted by the 'codec' package
func (ch *Chain) CallView(scName string, funName string, params ...interface{}) (dict.Dict, error) {
	ch.Log.Infof("callView: %s::%s", scName, funName)
	ret, err := ch.callViewFull(NewCall(scName, funName, params...))
	if err != nil {
		ch.Log.Errorf("callView: %s::%s: %v", scName, funName, err)
		return nil, err
	}
	return ret, nil
}

// WaitForEmptyBacklog waits until the backlog queue of the chain becomes empty.
// It is useful when smart contract(s) in the test are posting asynchronous requests
// between chains.
//
// The call is needed in order to prevent finishing the test before all
// asynchronous request between chains are processed.
// Otherwise waiting is not necessary because all PostRequest calls by the test itself
// are synchronous and are processed immediately
func (ch *Chain) WaitForEmptyBacklog(maxWait ...time.Duration) {
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
				break
			}
		} else {
			time.Sleep(10 * time.Millisecond)
			if ch.backlogLen() == 0 {
				break
			}
		}
	}
}
