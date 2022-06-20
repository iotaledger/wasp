// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"fmt"
	"math"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/models/trie_blake2b"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/chain/mempool"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type CallParams struct {
	targetName string
	target     iscp.Hname
	epName     string
	entryPoint iscp.Hname
	ftokens    *iscp.FungibleTokens // ignored off-ledger
	nft        *iscp.NFT
	allowance  *iscp.Allowance
	gasBudget  uint64
	nonce      uint64 // ignored for on-ledger
	params     dict.Dict
	sender     iotago.Address
}

// NewCallParams creates structure which wraps in one object call parameters, used in PostRequestSync and callViewFull
// calls:
//  - 'scName' is a a name of the target smart contract
//  - 'funName' is a name of the target entry point (the function) of he smart contract program
//  - 'params' is either a dict.Dict, or a sequence of pairs 'paramName', 'paramValue' which constitute call parameters
//     The 'paramName' must be a string and 'paramValue' must different types (encoded based on type)
// With the WithTransfers the CallParams structure may be complemented with attached ftokens
// sent together with the request
func NewCallParams(scName, funName string, params ...interface{}) *CallParams {
	return NewCallParamsFromDict(scName, funName, parseParams(params))
}

func NewCallParamsFromDict(scName, funName string, par dict.Dict) *CallParams {
	ret := NewCallParamsFromDictByHname(iscp.Hn(scName), iscp.Hn(funName), par)
	ret.targetName = scName
	ret.epName = funName
	return ret
}

func NewCallParamsFromDictByHname(hContract, hFunction iscp.Hname, par dict.Dict) *CallParams {
	ret := &CallParams{
		target:     hContract,
		entryPoint: hFunction,
	}
	ret.params = dict.New()
	for k, v := range par {
		ret.params.Set(k, v)
	}
	return ret
}

func (r *CallParams) WithAllowance(allowance *iscp.Allowance) *CallParams {
	r.allowance = allowance.Clone()
	return r
}

func (r *CallParams) AddAllowance(allowance *iscp.Allowance) *CallParams {
	if r.allowance == nil {
		r.allowance = allowance.Clone()
	} else {
		r.allowance.Add(allowance)
	}
	return r
}

func (r *CallParams) AddAllowanceIotas(amount uint64) *CallParams {
	return r.AddAllowance(iscp.NewAllowance(amount, nil, nil))
}

func (r *CallParams) AddAllowanceNativeTokensVect(tokens ...*iotago.NativeToken) *CallParams {
	if r.allowance == nil {
		r.allowance = iscp.NewEmptyAllowance()
	}
	r.allowance.Assets.Add(&iscp.FungibleTokens{
		Tokens: tokens,
	})
	return r
}

func (r *CallParams) AddAllowanceNativeTokens(id *iotago.NativeTokenID, amount interface{}) *CallParams {
	if r.allowance == nil {
		r.allowance = iscp.NewEmptyAllowance()
	}
	r.allowance.Assets.Add(&iscp.FungibleTokens{
		Tokens: iotago.NativeTokens{&iotago.NativeToken{
			ID:     *id,
			Amount: util.ToBigInt(amount),
		}},
	})
	return r
}

func (r *CallParams) AddAllowanceNFTs(nfts ...iotago.NFTID) *CallParams {
	return r.AddAllowance(iscp.NewAllowance(0, nil, nfts))
}

func (r *CallParams) WithFungibleTokens(assets *iscp.FungibleTokens) *CallParams {
	if r.allowance == nil {
		r.allowance = iscp.NewEmptyAllowance()
	}
	r.ftokens = assets.Clone()
	return r
}

func (r *CallParams) AddFungibleTokens(assets *iscp.FungibleTokens) *CallParams {
	if r.ftokens == nil {
		r.ftokens = assets.Clone()
	} else {
		r.ftokens.Add(assets)
	}
	return r
}

func (r *CallParams) AddIotas(amount uint64) *CallParams {
	return r.AddFungibleTokens(iscp.NewFungibleTokens(amount, nil))
}

func (r *CallParams) AddNativeTokensVect(tokens ...*iotago.NativeToken) *CallParams {
	return r.AddFungibleTokens(&iscp.FungibleTokens{
		Tokens: tokens,
	})
}

func (r *CallParams) AddNativeTokens(tokenID *iotago.NativeTokenID, amount interface{}) *CallParams {
	return r.AddFungibleTokens(&iscp.FungibleTokens{
		Tokens: iotago.NativeTokens{&iotago.NativeToken{
			ID:     *tokenID,
			Amount: util.ToBigInt(amount),
		}},
	})
}

// Adds an nft to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) WithNFT(nft *iscp.NFT) *CallParams {
	r.nft = nft
	return r
}

func (r *CallParams) GasBudget() uint64 {
	return r.gasBudget
}

func (r *CallParams) WithGasBudget(gasBudget uint64) *CallParams {
	r.gasBudget = gasBudget
	return r
}

func (r *CallParams) WithMaxAffordableGasBudget() *CallParams {
	r.gasBudget = math.MaxUint64
	return r
}

func (r *CallParams) WithNonce(nonce uint64) *CallParams {
	r.nonce = nonce
	return r
}

func (r *CallParams) WithSender(sender iotago.Address) *CallParams {
	r.sender = sender
	return r
}

// NewRequestOffLedger creates off-ledger request from parameters
func (r *CallParams) NewRequestOffLedger(chainID *iscp.ChainID, keyPair *cryptolib.KeyPair) iscp.OffLedgerRequest {
	ret := iscp.NewOffLedgerRequest(chainID, r.target, r.entryPoint, r.params, r.nonce).
		WithGasBudget(r.gasBudget).
		WithAllowance(r.allowance)
	return ret.Sign(keyPair)
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
		var key string
		switch p := params[2*i].(type) {
		case string:
			key = p
		case kv.Key:
			key = string(p)
		default:
			panic("WithParams: string or kv.Key expected")
		}
		par[key] = params[2*i+1]
	}
	return par
}

func (ch *Chain) createRequestTx(req *CallParams, keyPair *cryptolib.KeyPair) (*iotago.Transaction, error) {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	L1Iotas := ch.Env.L1Iotas(keyPair.Address())
	if L1Iotas == 0 {
		return nil, fmt.Errorf("PostRequestSync - Signer doesn't own any iotas on L1")
	}
	addr := keyPair.Address()
	allOuts, allOutIDs := ch.Env.utxoDB.GetUnspentOutputs(addr)

	sender := req.sender
	if sender == nil {
		sender = keyPair.Address()
	}

	tx, err := transaction.NewRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair:    keyPair,
		SenderAddress:    sender,
		UnspentOutputs:   allOuts,
		UnspentOutputIDs: allOutIDs,
		Request: &iscp.RequestParameters{
			TargetAddress:  ch.ChainID.AsAddress(),
			FungibleTokens: req.ftokens,
			Metadata: &iscp.SendMetadata{
				TargetContract: req.target,
				EntryPoint:     req.entryPoint,
				Params:         req.params,
				Allowance:      req.allowance,
				GasBudget:      req.gasBudget,
			},
			Options: iscp.SendOptions{},
		},
		NFT:                          req.nft,
		DisableAutoAdjustDustDeposit: ch.Env.disableAutoAdjustDustDeposit,
	})
	if err != nil {
		return nil, err
	}

	if tx.Essence.Outputs[0].Deposit() == 0 {
		return nil, xerrors.New("createRequestTx: amount == 0. Consider: solo.InitOptions{AutoAdjustDustDeposit: true}")
	}
	return tx, err
}

// requestFromParams creates an on-ledger request without posting the transaction. It is intended
// mainly for estimating gas.
func (ch *Chain) requestFromParams(req *CallParams, keyPair *cryptolib.KeyPair) (iscp.Request, error) {
	ch.Env.ledgerMutex.Lock()
	defer ch.Env.ledgerMutex.Unlock()

	tx, err := ch.createRequestTx(req, keyPair)
	if err != nil {
		return nil, err
	}
	reqs, err := iscp.RequestsInTransaction(tx)
	require.NoError(ch.Env.T, err)

	for _, r := range reqs[*ch.ChainID] {
		// return the first one
		return r, nil
	}
	panic("unreachable")
}

// RequestFromParamsToLedger creates transaction with one request based on parameters and sigScheme
// Then it adds it to the ledger, atomically.
// Locking on the mutex is needed to prevent mess when several goroutines work on the same address
func (ch *Chain) RequestFromParamsToLedger(req *CallParams, keyPair *cryptolib.KeyPair) (*iotago.Transaction, iscp.RequestID, error) {
	ch.Env.ledgerMutex.Lock()
	defer ch.Env.ledgerMutex.Unlock()

	tx, err := ch.createRequestTx(req, keyPair)
	if err != nil {
		return nil, iscp.RequestID{}, err
	}
	err = ch.Env.AddToLedger(tx)
	// once we created transaction successfully, it should be added to the ledger smoothly
	require.NoError(ch.Env.T, err)
	txid, err := tx.ID()
	require.NoError(ch.Env.T, err)

	return tx, iscp.NewRequestID(txid, 0), nil
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
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	r := req.NewRequestOffLedger(ch.ChainID, keyPair)
	return ch.RunOffLedgerRequest(r)
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, keyPair *cryptolib.KeyPair) (*iotago.Transaction, dict.Dict, error) {
	tx, receipt, res, err := ch.PostRequestSyncExt(req, keyPair)
	if err != nil {
		return tx, res, err
	}
	return tx, res, receipt.Error.AsGoError()
}

// LastReceipt returns the receipt fot the latest request processed by the chain, will return nil if the last block is empty
func (ch *Chain) LastReceipt() *blocklog.RequestReceipt {
	lastBlockReceipts := ch.GetRequestReceiptsForBlock()
	if len(lastBlockReceipts) == 0 {
		return nil
	}
	return lastBlockReceipts[len(lastBlockReceipts)-1]
}

func (ch *Chain) checkCanAffordFee(fee uint64, req *CallParams, keyPair *cryptolib.KeyPair) error {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	agentID := iscp.NewAgentID(keyPair.GetPublicKey().AsEd25519Address())
	policy := ch.GetGasFeePolicy()
	available := uint64(0)
	if policy.GasFeeTokenID == nil {
		available = ch.L2Iotas(agentID)
		if req.ftokens != nil {
			available += req.ftokens.Iotas
		}
		if req.allowance != nil {
			available -= req.allowance.Assets.Iotas
		}
	} else {
		n := ch.L2NativeTokens(agentID, policy.GasFeeTokenID)
		if req.ftokens != nil {
			n.Add(n, req.ftokens.AmountNativeToken(policy.GasFeeTokenID))
		}
		if req.allowance != nil {
			n.Sub(n, req.allowance.Assets.AmountNativeToken(policy.GasFeeTokenID))
		}
		if n.IsUint64() {
			available = n.Uint64()
		} else {
			available = math.MaxUint64
		}
	}
	if available < fee {
		return fmt.Errorf("sender's available tokens on L2 (%d) is less than the %d required", available, fee)
	}
	return nil
}

func (ch *Chain) PostRequestSyncExt(req *CallParams, keyPair *cryptolib.KeyPair) (*iotago.Transaction, *blocklog.RequestReceipt, dict.Dict, error) {
	defer ch.logRequestLastBlock()

	tx, _, err := ch.RequestFromParamsToLedger(req, keyPair)
	require.NoError(ch.Env.T, err)
	reqs, err := ch.Env.RequestsForChain(tx, ch.ChainID)
	require.NoError(ch.Env.T, err)
	results := ch.runRequestsSync(reqs, "post")
	if len(results) == 0 {
		return nil, nil, nil, xerrors.New("request has been skipped")
	}
	res := results[0]
	return tx, res.Receipt, res.Return, res.Error
}

// EstimateGasOnLedger executes the given on-ledger request without committing
// any changes in the ledger. It returns the amount of gas consumed.
// if useFakeBalance is `true` the request will be executed as if the sender had enough iotas to cover the maximum gas allowed
// WARNING: Gas estimation is just an "estimate", there is no guarantees that the real call will bear the same cost, due to the turing-completeness of smart contracts
func (ch *Chain) EstimateGasOnLedger(req *CallParams, keyPair *cryptolib.KeyPair, useFakeBudget ...bool) (gas, gasFee uint64, err error) {
	if len(useFakeBudget) > 0 && useFakeBudget[0] {
		req.WithGasBudget(math.MaxUint64)
	}
	r, err := ch.requestFromParams(req, keyPair)
	if err != nil {
		return 0, 0, err
	}

	res := ch.estimateGas(r)

	return res.Receipt.GasBurned, res.Receipt.GasFeeCharged, res.Receipt.Error.AsGoError()
}

// EstimateGasOffLedger executes the given on-ledger request without committing
// any changes in the ledger. It returns the amount of gas consumed.
// if useFakeBalance is `true` the request will be executed as if the sender had enough iotas to cover the maximum gas allowed
// WARNING: Gas estimation is just an "estimate", there is no guarantees that the real call will bear the same cost, due to the turing-completeness of smart contracts
func (ch *Chain) EstimateGasOffLedger(req *CallParams, keyPair *cryptolib.KeyPair, useMaxBalance ...bool) (gas, gasFee uint64, err error) {
	if len(useMaxBalance) > 0 && useMaxBalance[0] {
		req.WithGasBudget(math.MaxUint64)
	}
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	r := req.NewRequestOffLedger(ch.ChainID, keyPair)
	res := ch.estimateGas(r)

	return res.Receipt.GasBurned, res.Receipt.GasFeeCharged, res.Receipt.Error.AsGoError()
}

func (ch *Chain) ResolveVMError(e *iscp.UnresolvedVMError) *iscp.VMError {
	resolved, err := errors.Resolve(e, func(contractName string, funcName string, params dict.Dict) (dict.Dict, error) {
		return ch.CallView(contractName, funcName, params)
	})
	require.NoError(ch.Env.T, err)
	return resolved
}

// CallView calls the view entry point of the smart contract.
// The call params should be either a dict.Dict, or pairs of ('paramName',
// 'paramValue') where 'paramName' is a string and 'paramValue' must be of type
// accepted by the 'codec' package
func (ch *Chain) CallView(scName, funName string, params ...interface{}) (dict.Dict, error) {
	ch.Log().Debugf("callView: %s::%s", scName, funName)
	return ch.CallViewByHname(iscp.Hn(scName), iscp.Hn(funName), params...)
}

func (ch *Chain) CallViewByHname(hContract, hFunction iscp.Hname, params ...interface{}) (dict.Dict, error) {
	ch.Log().Debugf("callView: %s::%s", hContract.String(), hFunction.String())

	p := parseParams(params)

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vmctx := viewcontext.New(ch)
	ch.StateReader.SetBaseline()
	return vmctx.CallViewExternal(hContract, hFunction, p)
}

// GetMerkleProofRaw returns Merkle proof of the key in the state
func (ch *Chain) GetMerkleProofRaw(key []byte) *trie_blake2b.Proof {
	ch.Log().Debugf("GetMerkleProof")

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vmctx := viewcontext.New(ch)
	ch.StateReader.SetBaseline()
	ret, err := vmctx.GetMerkleProof(key)
	require.NoError(ch.Env.T, err)
	return ret
}

// GetBlockProof returns Merkle proof of the key in the state
func (ch *Chain) GetBlockProof(blockIndex uint32) (*blocklog.BlockInfo, *trie_blake2b.Proof, error) {
	ch.Log().Debugf("GetBlockProof")

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vmctx := viewcontext.New(ch)
	ch.StateReader.SetBaseline()
	biBin, retProof, err := vmctx.GetBlockProof(blockIndex)
	if err != nil {
		return nil, nil, err
	}
	retBlockInfo, err := blocklog.BlockInfoFromBytes(blockIndex, biBin)
	if err != nil {
		return nil, nil, err
	}

	return retBlockInfo, retProof, nil
}

// GetMerkleProof return the merkle proof of the key in the smart contract. Assumes Merkle model is used
func (ch *Chain) GetMerkleProof(scHname iscp.Hname, key []byte) *trie_blake2b.Proof {
	return ch.GetMerkleProofRaw(kv.Concat(scHname, key))
}

// GetL1Commitment returns state commitment taken from the anchor output
func (ch *Chain) GetL1Commitment() *state.L1Commitment {
	anchorOutput := ch.GetAnchorOutput()
	ret, err := state.L1CommitmentFromAnchorOutput(anchorOutput.GetAliasOutput())
	require.NoError(ch.Env.T, err)
	return &ret
}

// GetRootCommitment calculates root commitment from state
func (ch *Chain) GetRootCommitment() trie.VCommitment {
	vmctx := viewcontext.New(ch)
	ch.StateReader.SetBaseline()
	ret, err := vmctx.GetRootCommitment()
	require.NoError(ch.Env.T, err)
	return ret
}

// GetContractStateCommitment returns commitment to the state of the specific contract, if possible
func (ch *Chain) GetContractStateCommitment(hn iscp.Hname) ([]byte, error) {
	vmctx := viewcontext.New(ch)
	ch.StateReader.SetBaseline()
	return vmctx.GetContractStateCommitment(hn)
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
		mstats := ch.MempoolInfo()
		if p(mstats) {
			return true
		}
		if time.Now().After(deadline) {
			ch.Log().Errorf("WaitUntil failed waiting max %v", maxw)
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

const waitUntilMempoolIsEmptyDefaultTimeout = 5 * time.Second

func (ch *Chain) WaitUntilMempoolIsEmpty(timeout ...time.Duration) bool {
	realTimeout := waitUntilMempoolIsEmptyDefaultTimeout
	if len(timeout) > 0 {
		realTimeout = timeout[0]
	}
	startTime := time.Now()
	ret := ch.mempool.WaitInBufferEmpty(timeout...)
	if !ret {
		return false
	}
	remainingTimeout := realTimeout - time.Since(startTime)
	return ch.mempool.WaitPoolEmpty(remainingTimeout)
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
	return ch.mempool.Info(ch.Env.GlobalTime())
}
