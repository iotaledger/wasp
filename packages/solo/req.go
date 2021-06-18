// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type CallParams struct {
	targetName  string
	target      coretypes.Hname
	epName      string
	entryPoint  coretypes.Hname
	transfer    *ledgerstate.ColoredBalances
	mintAmount  uint64
	mintAddress ledgerstate.Address
	args        requestargs.RequestArgs
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
//  - 'params' is either a dict.Dict, or a sequence of pairs 'paramName', 'paramValue' which constitute call parameters
//     The 'paramName' must be a string and 'paramValue' must different types (encoded based on type)
// With the WithTransfers the CallParams structure may be complemented with attached colored
// tokens sent together with the request
func NewCallParams(scName, funName string, params ...interface{}) *CallParams {
	return NewCallParamsFromDic(scName, funName, parseParams(params))
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
	var retOptimized map[kv.Key][]byte
	ret.args, retOptimized = requestargs.NewOptimizedRequestArgs(parseParams(params), optSize)
	return ret, retOptimized
}

// WithTransfer is a shorthand for the most often used case where only
// a single color is transferred by WithTransfers
func (r *CallParams) WithTransfer(color ledgerstate.Color, amount uint64) *CallParams {
	return r.WithTransfers(map[ledgerstate.Color]uint64{color: amount})
}

// WithTransfers complement CallParams structure with the colored balances of tokens
// in the form of a collection of pairs 'color': 'balance'
func (r *CallParams) WithTransfers(transfer map[ledgerstate.Color]uint64) *CallParams {
	r.transfer = ledgerstate.NewColoredBalances(transfer)
	return r
}

func (r *CallParams) WithIotas(amount uint64) *CallParams {
	return r.WithTransfer(ledgerstate.ColorIOTA, amount)
}

// WithMint adds additional mint proof
func (r *CallParams) WithMint(targetAddress ledgerstate.Address, amount uint64) *CallParams {
	r.mintAddress = targetAddress
	r.mintAmount = amount
	return r
}

// NewRequestOffLedger creates off-ledger request from parameters
func (r *CallParams) NewRequestOffLedger(keyPair *ed25519.KeyPair) *request.RequestOffLedger {
	ret := request.NewRequestOffLedger(r.target, r.entryPoint, r.args).WithTransfer(r.transfer)
	ret.Sign(keyPair)
	return ret
}

func parseParams(params []interface{}) dict.Dict {
	if len(params) == 1 {
		return params[0].(dict.Dict)
	}
	return codec.MakeDict(toMap(params))
}

// makes map without hashing
func toMap(params []interface{}) map[string]interface{} {
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
func (ch *Chain) RequestFromParamsToLedger(req *CallParams, keyPair *ed25519.KeyPair) (*ledgerstate.Transaction, error) {
	if req.transfer == nil || req.transfer.Size() == 0 {
		return nil, xerrors.New("transfer can't be empty")
	}

	ch.Env.ledgerMutex.Lock()
	defer ch.Env.ledgerMutex.Unlock()

	if keyPair == nil {
		keyPair = ch.OriginatorKeyPair
	}
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	allOuts := ch.Env.utxoDB.GetAddressOutputs(addr)

	metadata := request.NewRequestMetadata().
		WithTarget(req.target).
		WithEntryPoint(req.entryPoint).
		WithArgs(req.args)

	data := metadata.Bytes()
	mdataBack := request.RequestMetadataFromBytes(data)
	require.True(ch.Env.T, mdataBack.ParsedOk())

	txb := utxoutil.NewBuilder(allOuts...).WithTimestamp(ch.Env.LogicalTime())
	var err error
	err = txb.AddExtendedOutputConsume(ch.ChainID.AsAddress(), data, req.transfer.Map())
	require.NoError(ch.Env.T, err)
	if req.mintAmount > 0 {
		err = txb.AddMintingOutputConsume(req.mintAddress, req.mintAmount)
		require.NoError(ch.Env.T, err)
	}

	err = txb.AddRemainderOutputIfNeeded(addr, nil, true)
	require.NoError(ch.Env.T, err)

	tx, err := txb.BuildWithED25519(keyPair)
	require.NoError(ch.Env.T, err)

	err = ch.Env.AddToLedger(tx)
	require.NoError(ch.Env.T, err)
	return tx, nil
}

// PostRequestSync posts a request synchronously  sent by the test program to the smart contract on the same or another chain:
//  - creates a request transaction with the request block on it. The sigScheme is used to
//    sign the inputs of the transaction or OriginatorKeyPair is used if parameter is nil
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
func (ch *Chain) PostRequestSync(req *CallParams, keyPair *ed25519.KeyPair) (dict.Dict, error) {
	_, ret, err := ch.PostRequestSyncTx(req, keyPair)
	return ret, err
}

func (ch *Chain) PostRequestOffLedger(req *CallParams, keyPair *ed25519.KeyPair) (dict.Dict, error) {
	if keyPair == nil {
		keyPair = ch.OriginatorKeyPair
	}
	r := req.NewRequestOffLedger(keyPair)
	return ch.runRequestsSync([]coretypes.Request{r}, "off-ledger")
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, keyPair *ed25519.KeyPair) (*ledgerstate.Transaction, dict.Dict, error) {
	tx, err := ch.RequestFromParamsToLedger(req, keyPair)
	if err != nil {
		return nil, nil, err
	}
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	res, err := ch.runRequestsSync(reqs, "post")
	if err != nil {
		return nil, nil, err
	}

	return tx, res, nil
}

// callViewFull calls the view entry point of the smart contract
// with params wrapped into the CallParams object. The transfer part, fs any, is ignored
//nolint:unused
func (ch *Chain) callViewFull(req *CallParams) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.StateReader, ch.proc, ch.Log)
	a, ok, err := req.args.SolidifyRequestArguments(ch.Env.blobCache)
	if err != nil || !ok {
		return nil, fmt.Errorf("solo.internal error: can't solidify args")
	}
	return vctx.CallView(req.target, req.entryPoint, a)
}

// CallView calls the view entry point of the smart contract.
// The call params should be either a dict.Dict, or pairs of ('paramName',
// 'paramValue') where 'paramName' is a string and 'paramValue' must be of type
// accepted by the 'codec' package
func (ch *Chain) CallView(scName, funName string, params ...interface{}) (dict.Dict, error) {
	ch.Log.Infof("callView: %s::%s", scName, funName)

	p := parseParams(params)

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.StateReader, ch.proc, ch.Log)
	ch.StateReader.SetBaseline()
	return vctx.CallView(coretypes.Hn(scName), coretypes.Hn(funName), p)
}

// WaitForRequestsThrough waits for the moment when counters for incoming requests and removed
// requests in the mempool of the chain both become equal to the specified number
func (ch *Chain) WaitForRequestsThrough(numReq int, maxWait ...time.Duration) bool {
	maxw := 5 * time.Second
	var deadline time.Time
	if len(maxWait) > 0 {
		maxw = maxWait[0]
	}
	deadline = time.Now().Add(maxw)
	for {
		mstats := ch.mempool.Stats()
		if mstats.InBufCounter == numReq && mstats.OutPoolCounter == numReq {
			return true
		}
		if time.Now().After(deadline) {
			ch.Log.Errorf("WaitForRequestsThrough. failed waiting max %v for %d requests through . Current IN: %d, OUT: %d",
				maxw, numReq, mstats.InBufCounter, mstats.OutPoolCounter)
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}
