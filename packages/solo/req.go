// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"time"

	"github.com/iotaledger/wasp/packages/transaction"

	"github.com/iotaledger/wasp/packages/chain/mempool"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type CallParams struct {
	targetName string
	target     iscp.Hname
	epName     string
	entryPoint iscp.Hname
	assets     *iscp.Assets // ignored off-ledger
	transfer   *iscp.Assets
	gasBudget  uint64
	nonce      uint64 // ignored for on-ledger
	params     dict.Dict
}

func NewCallParamsFromDic(scName, funName string, par dict.Dict) *CallParams {
	ret := &CallParams{
		targetName: scName,
		target:     iscp.Hn(scName),
		epName:     funName,
		entryPoint: iscp.Hn(funName),
	}
	ret.params = dict.New()
	for k, v := range par {
		ret.params.Set(k, v)
	}
	return ret
}

// NewCallParams creates structure which wraps in one object call parameters, used in PostRequestSync and callViewFull
// calls:
//  - 'scName' is a a name of the target smart contract
//  - 'funName' is a name of the target entry point (the function) of he smart contract program
//  - 'params' is either a dict.Dict, or a sequence of pairs 'paramName', 'paramValue' which constitute call parameters
//     The 'paramName' must be a string and 'paramValue' must different types (encoded based on type)
// With the WithTransfers the CallParams structure may be complemented with attached assets
// sent together with the request
func NewCallParams(scName, funName string, params ...interface{}) *CallParams {
	return NewCallParamsFromDic(scName, funName, parseParams(params))
}

func (r *CallParams) WithTransfer(transfer *iscp.Assets) *CallParams {
	r.transfer = transfer
	return r
}

func (r *CallParams) WithAssets(assets *iscp.Assets) *CallParams {
	r.assets = assets
	return r
}

func (r *CallParams) WithGasBudget(gasBudget uint64) *CallParams {
	r.gasBudget = gasBudget
	return r
}

func (r *CallParams) WithNonce(nonce uint64) *CallParams {
	r.nonce = nonce
	return r
}

func (r *CallParams) WithIotas(amount uint64) *CallParams {
	return r.WithTransfer(iscp.NewAssets(amount, nil))
}

// NewRequestOffLedger creates off-ledger request from parameters
func (r *CallParams) NewRequestOffLedger(chainID *iscp.ChainID, keyPair *cryptolib.KeyPair) *iscp.OffLedgerRequestData {
	ret := iscp.NewOffLedgerRequest(chainID, r.target, r.entryPoint, r.params, 0)
	ret.Sign(*keyPair)
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
// Locking on the mutex is needed to prevent mess when several goroutines work on the same address
func (ch *Chain) RequestFromParamsToLedger(req *CallParams, keyPair *cryptolib.KeyPair) (*iotago.Transaction, iscp.RequestID, error) {
	ch.Env.ledgerMutex.Lock()
	defer ch.Env.ledgerMutex.Unlock()

	if keyPair == nil {
		keyPair = &ch.OriginatorPrivateKey
	}
	addr := iotago.Ed25519AddressFromPubKey(keyPair.PublicKey)
	allOuts, ids := ch.Env.utxoDB.GetUnspentOutputs(&addr)

	tx, err := transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    *keyPair,
		UnspentOutputs:   allOuts,
		UnspentOutputIDs: ids,
		Requests: []*iscp.RequestParameters{{
			TargetAddress: ch.ChainID.AsAddress(),
			Assets:        req.assets,
			Metadata: &iscp.SendMetadata{
				TargetContract: req.target,
				EntryPoint:     req.entryPoint,
				Params:         req.params,
				Transfer:       req.transfer,
				GasBudget:      req.gasBudget,
			},
			Options: nil,
		}},
		RentStructure: ch.Env.utxoDB.RentStructure(),
	})
	require.NoError(ch.Env.T, err)

	err = ch.Env.AddToLedger(tx)
	require.NoError(ch.Env.T, err)
	txid, err := tx.ID()
	require.NoError(ch.Env.T, err)

	return tx, iscp.NewRequestID(*txid, 0), nil
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
func (ch *Chain) PostRequestSync(req *CallParams, keyPair *cryptolib.KeyPair) (dict.Dict, error) {
	_, ret, err := ch.PostRequestSyncTx(req, keyPair)
	return ret, err
}

func (ch *Chain) PostRequestOffLedger(req *CallParams, keyPair *cryptolib.KeyPair) (dict.Dict, error) {
	defer ch.logRequestLastBlock()

	if keyPair == nil {
		keyPair = &ch.OriginatorPrivateKey
	}
	r := req.NewRequestOffLedger(ch.ChainID, keyPair)
	res, err := ch.runRequestsSync([]iscp.RequestData{r}, "off-ledger")
	if err != nil {
		return nil, err
	}
	return res, ch.mustGetErrorFromReceipt(r.ID())
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, keyPair *cryptolib.KeyPair) (*iotago.Transaction, dict.Dict, error) {
	defer ch.logRequestLastBlock()

	tx, reqid, err := ch.RequestFromParamsToLedger(req, keyPair)
	if err != nil {
		return tx, nil, err
	}
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	res, err := ch.runRequestsSync(reqs, "post")
	if err != nil {
		return tx, nil, err
	}
	return tx, res, ch.mustGetErrorFromReceipt(reqid)
}

func (ch *Chain) mustGetErrorFromReceipt(reqid iscp.RequestID) error {
	rec, _, _, ok := ch.GetRequestReceipt(reqid)
	require.True(ch.Env.T, ok)
	var err error
	if len(rec.Error) > 0 {
		err = xerrors.New(rec.Error)
	}
	return err
}

// callViewFull calls the view entry point of the smart contract
// with params wrapped into the CallParams object. The transfer part, fs any, is ignored
//nolint:unused
func (ch *Chain) callViewFull(req *CallParams) (dict.Dict, error) {
	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vctx := viewcontext.New(ch.ChainID, ch.StateReader, ch.proc, ch.Log)
	return vctx.CallView(req.target, req.entryPoint, req.params)
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
	return vctx.CallView(iscp.Hn(scName), iscp.Hn(funName), p)
}

// WaitUntil waits until the condition specified by the given predicate yields true
func (ch *Chain) WaitUntil(p func(mempool.MempoolInfo) bool, maxWait ...time.Duration) bool {
	maxw := 10 * time.Second
	var deadline time.Time
	if len(maxWait) > 0 {
		maxw = maxWait[0]
	}
	deadline = time.Now().Add(maxw)
	for {
		mstats := ch.mempool.Info()
		if p(mstats) {
			return true
		}
		if time.Now().After(deadline) {
			ch.Log.Errorf("WaitUntil failed waiting max %v", maxw)
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// WaitForRequestsThrough waits for the moment when counters for incoming requests and removed
// requests in the mempool of the chain both become equal to the specified number
func (ch *Chain) WaitForRequestsThrough(numReq int, maxWait ...time.Duration) bool {
	return ch.WaitUntil(func(mstats mempool.MempoolInfo) bool {
		return mstats.InBufCounter == numReq && mstats.OutPoolCounter == numReq
	}, maxWait...)
}

// MempoolInfo returns stats about the chain mempool
func (ch *Chain) MempoolInfo() mempool.MempoolInfo {
	return ch.mempool.Info()
}
