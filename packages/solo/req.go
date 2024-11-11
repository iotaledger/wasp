// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	vmerrors "github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

type CallParams struct {
	msg       isc.Message
	assets    *isc.Assets // ignored off-ledger
	allowance *isc.Assets
	gasBudget uint64
	nonce     uint64 // ignored for on-ledger
	sender    *cryptolib.Address
}

// NewCallParams creates a structure that wraps in one object call parameters,
// used in PostRequestSync and CallView
func NewCallParams(msg isc.Message, contract ...any) *CallParams {
	p := &CallParams{
		msg:       msg,
		assets:    isc.NewEmptyAssets(),
		allowance: isc.NewEmptyAssets(),
	}

	if len(contract) > 0 {
		p = p.WithTargetContract(contract[0])
	}

	return p
}

// NewCallParamsEx is a shortcut for NewCallParams
func NewCallParamsEx(c, ep string, params ...isc.CallArguments) *CallParams {
	return NewCallParams(isc.NewMessageFromNames(c, ep, params...))
}

func (r *CallParams) WithTargetContract(contract any) *CallParams {
	switch contract := contract.(type) {
	case string:
		r.msg.Target.Contract = isc.Hn(contract)
	case isc.Hname:
		r.msg.Target.Contract = contract
	default:
		panic(fmt.Sprintf("unsupported contract type %T", contract))
	}

	return r
}

func (r *CallParams) WithAllowance(allowance *isc.Assets) *CallParams {
	r.allowance = allowance.Clone()
	return r
}

func (r *CallParams) AddAllowance(allowance *isc.Assets) *CallParams {
	if r.allowance == nil {
		r.allowance = allowance.Clone()
	} else {
		r.allowance.Add(allowance)
	}
	return r
}

func (r *CallParams) AddAllowanceBaseTokens(amount coin.Value) *CallParams {
	return r.AddAllowanceCoins(coin.BaseTokenType, amount)
}

// func (r *CallParams) AddAllowanceNativeTokensVect(nativeTokens ...*iotago.NativeToken) *CallParams {
// 	if r.allowance == nil {
// 		r.allowance = isc.NewEmptyAssets()
// 	}

// 	r.allowance.Add(&isc.Assets{
// 		NativeTokens: nativeTokens,
// 	})
// 	return r
// }

func (r *CallParams) AddAllowanceCoins(coinType coin.Type, amount coin.Value) *CallParams {
	if r.allowance == nil {
		r.allowance = isc.NewEmptyAssets()
	}

	r.allowance.AddCoin(coinType, amount)

	return r
}

func (r *CallParams) AddAllowanceNFTs(nftIDs ...iotago.ObjectID) *CallParams {
	if r.allowance == nil {
		r.allowance = isc.NewEmptyAssets()
	}

	for _, nftId := range nftIDs {
		r.allowance.AddObject(nftId)
	}

	return r
}

func (r *CallParams) WithAssets(assets *isc.Assets) *CallParams {
	r.assets = assets.Clone()
	return r
}

func (r *CallParams) WithFungibleTokens(ftokens isc.CoinBalances) *CallParams {
	r.assets.Coins = ftokens.Clone()
	return r
}

func (r *CallParams) AddFungibleTokens(ftokens isc.CoinBalances) *CallParams {
	r.assets.Add(ftokens.ToAssets())
	return r
}

func (r *CallParams) AddBaseTokens(amount coin.Value) *CallParams {
	r.assets.AddBaseTokens(amount)
	return r
}

func (r *CallParams) AddCoin(coinType coin.Type, amount coin.Value) *CallParams {
	r.assets.AddCoin(coinType, amount)
	return r
}

// Adds an nft to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) WithObject(objectID iotago.ObjectID) *CallParams {
	r.assets.AddObject(objectID)
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

func (r *CallParams) WithSender(sender *cryptolib.Address) *CallParams {
	r.sender = sender
	return r
}

// NewRequestOffLedger creates off-ledger request from parameters
func (r *CallParams) NewRequestOffLedger(ch *Chain, keyPair *cryptolib.KeyPair) isc.OffLedgerRequest {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	if r.nonce == 0 {
		r.nonce = ch.Nonce(isc.NewAddressAgentID(keyPair.Address()))
	}
	ret := isc.NewOffLedgerRequest(ch.ID(), r.msg, r.nonce, r.gasBudget).
		WithAllowance(r.allowance)
	return ret.Sign(keyPair)
}

func (r *CallParams) NewRequestImpersonatedOffLedger(ch *Chain, address *cryptolib.Address) isc.OffLedgerRequest {
	if r.nonce == 0 {
		r.nonce = ch.Nonce(isc.NewAddressAgentID(address))
	}
	ret := isc.NewOffLedgerRequest(ch.ID(), r.msg, r.nonce, r.gasBudget).
		WithAllowance(r.allowance)

	return isc.NewImpersonatedOffLedgerRequest(ret.(*isc.OffLedgerRequestData)).WithSenderAddress(address)
}

// requestFromParams creates an on-ledger request without sending it to L1. It is intended
// mainly for estimating gas.
func (ch *Chain) requestFromParams(cp *CallParams, keyPair *cryptolib.KeyPair) (isc.Request, error) {
	panic("TODO")
}

func (env *Solo) makeBaseTokenCoinsWithExactly(
	keyPair *cryptolib.KeyPair,
	values ...coin.Value,
) []*iotago.ObjectRef {
	allCoins := env.L1BaseTokenCoins(keyPair.Address())
	require.NotEmpty(env.T, allCoins)
	coinToSplit, ok := lo.Find(allCoins, func(item *iotajsonrpc.Coin) bool {
		return item.Balance.Uint64() >= uint64(lo.Sum(values))+iotaclient.DefaultGasBudget
	})
	require.True(env.T, ok, "not enough base tokens")
	tx := lo.Must(env.IotaClient().PayIota(
		env.ctx,
		iotaclient.PayIotaRequest{
			Signer:     keyPair.Address().AsIotaAddress(),
			InputCoins: []*iotago.ObjectID{coinToSplit.CoinObjectID},
			Amount: lo.Map(values, func(v coin.Value, _ int) *iotajsonrpc.BigInt {
				return iotajsonrpc.NewBigInt(uint64(v))
			}),
			Recipients: lo.Map(values, func(v coin.Value, _ int) *iotago.Address {
				return keyPair.Address().AsIotaAddress()
			}),
			GasBudget: iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	))
	txnResponse, err := env.IotaClient().SignAndExecuteTransaction(
		env.ctx,
		cryptolib.SignerToIotaSigner(keyPair),
		tx.TxBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(env.T, err)
	require.True(env.T, txnResponse.Effects.Data.IsSuccess())
	allCoins = env.L1BaseTokenCoins(keyPair.Address())
	return lo.Map(values, func(v coin.Value, _ int) *iotago.ObjectRef {
		c, i, ok := lo.FindIndexOf(allCoins, func(coin *iotajsonrpc.Coin) bool {
			return coin.Balance.Uint64() == uint64(v)
		})
		require.True(env.T, ok)
		allCoins = lo.DropByIndex(allCoins, i)
		return c.Ref()
	})
}

// RequestFromParamsToLedger creates transaction with one request based on parameters and sigScheme
// Then it adds it to the ledger, atomically.
// Locking on the mutex is needed to prevent mess when several goroutines work on the same address
func (ch *Chain) RequestFromParamsToLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.RequestID, *iotajsonrpc.IotaTransactionBlockResponse, error) {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	res, err := ch.Env.ISCMoveClient().CreateAndSendRequestWithAssets(
		ch.Env.ctx,
		keyPair,
		ch.Env.ISCPackageID(),
		ch.ID().AsAddress().AsIotaAddress(),
		req.assets.AsISCMove(),
		&iscmove.Message{
			Contract: uint32(req.msg.Target.Contract),
			Function: uint32(req.msg.Target.EntryPoint),
			Args:     req.msg.Params,
		},
		req.allowance.AsISCMove(),
		req.gasBudget,
		nil,
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	if err != nil {
		return isc.RequestID{}, nil, err
	}
	reqRef, err := res.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	if err != nil {
		return isc.RequestID{}, nil, err
	}

	return isc.RequestID(*reqRef.ObjectID), res, nil
}

// PostRequestSync posts a request synchronously sent by the test program to
// the smart contract on the same or another chain.
//
// Note that in a real network of Wasp nodes, posting a transaction is completely
// asynchronous, i.e. result of the call is not available to the originator of the post.
// Instead, the Solo environment makes PostRequestSync a synchronous call,
// making it possible to step-by-step debug the smart contract logic.
func (ch *Chain) PostRequestSync(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, error) {
	_, _, res, err := ch.PostRequestSyncTx(req, keyPair)
	if err != nil {
		return nil, err
	}
	return res.Return, nil
}

func (ch *Chain) PostRequestOffLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, error) {
	r := req.NewRequestOffLedger(ch, keyPair)
	return ch.RunOffLedgerRequest(r)
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, keyPair *cryptolib.KeyPair) (
	isc.RequestID,
	*iotajsonrpc.IotaTransactionBlockResponse,
	*vm.RequestResult,
	error,
) {
	reqID, l1Res, vmRes, err := ch.PostRequestSyncExt(req, keyPair)
	if err != nil {
		return reqID, l1Res, vmRes, err
	}
	return reqID, l1Res, vmRes, ch.ResolveVMError(vmRes.Receipt.Error).AsGoError()
}

// LastReceipt returns the receipt for the latest request processed by the chain, will return nil if the last block is empty
func (ch *Chain) LastReceipt() *isc.Receipt {
	lastBlockReceipts := ch.GetRequestReceiptsForBlock()
	if len(lastBlockReceipts) == 0 {
		return nil
	}
	blocklogReceipt := lastBlockReceipts[len(lastBlockReceipts)-1]
	return blocklogReceipt.ToISCReceipt(ch.ResolveVMError(blocklogReceipt.Error))
}

func (ch *Chain) PostRequestSyncExt(
	callParams *CallParams,
	keyPair *cryptolib.KeyPair,
) (
	isc.RequestID,
	*iotajsonrpc.IotaTransactionBlockResponse,
	*vm.RequestResult,
	error,
) {
	if keyPair == nil {
		keyPair = ch.OriginatorPrivateKey
	}
	defer ch.logRequestLastBlock()

	reqID, l1Res, err := ch.RequestFromParamsToLedger(callParams, keyPair)
	require.NoError(ch.Env.T, err)

	reqWithObj, err := ch.Env.ISCMoveClient().GetRequestFromObjectID(ch.Env.ctx, (*iotago.ObjectID)(&reqID))
	req, err := isc.OnLedgerFromRequest(reqWithObj, keyPair.Address())

	results := ch.RunRequestsSync([]isc.Request{req}, "post")
	if len(results) == 0 {
		return isc.RequestID{}, nil, nil, errors.New("request has been skipped")
	}
	vmRes := results[0]

	return reqID, l1Res, vmRes, nil
}

// EstimateGasOnLedger executes the given on-ledger request without committing
// any changes in the ledger. It returns the amount of gas consumed.
// WARNING: Gas estimation is just an "estimate", there is no guarantees that the real call will bear the same cost, due to the turing-completeness of smart contracts
// TODO only a senderAddr, not a keyPair should be necessary to estimate (it definitely shouldn't fallback to the chain originator)
func (ch *Chain) EstimateGasOnLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, *blocklog.RequestReceipt, error) {
	reqCopy := *req
	r, err := ch.requestFromParams(&reqCopy, keyPair)
	if err != nil {
		return nil, nil, err
	}

	res := ch.estimateGas(r)

	return res.Return, res.Receipt, ch.ResolveVMError(res.Receipt.Error).AsGoError()
}

// EstimateGasOffLedger executes the given on-ledger request without committing
// any changes in the ledger. It returns the amount of gas consumed.
// WARNING: Gas estimation is just an "estimate", there is no guarantees that the real call will bear the same cost, due to the turing-completeness of smart contracts
func (ch *Chain) EstimateGasOffLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, *blocklog.RequestReceipt, error) {
	reqCopy := *req
	r := reqCopy.NewRequestImpersonatedOffLedger(ch, keyPair.Address())
	res := ch.estimateGas(r)
	return res.Return, res.Receipt, ch.ResolveVMError(res.Receipt.Error).AsGoError()
}

func (ch *Chain) ResolveVMError(e *isc.UnresolvedVMError) *isc.VMError {
	resolved, err := vmerrors.Resolve(e, ch.CallView)
	require.NoError(ch.Env.T, err)
	return resolved
}

// CallView calls a view entry point of a smart contract.
func (ch *Chain) CallView(msg isc.Message) (isc.CallArguments, error) {
	latestState, err := ch.LatestState()
	if err != nil {
		return nil, err
	}
	return ch.CallViewAtState(latestState, msg)
}

func (ch *Chain) CallViewWithContract(contract any, msg isc.Message) (isc.CallArguments, error) {
	switch contract := contract.(type) {
	case string:
		msg.Target.Contract = isc.Hn(contract)
	case isc.Hname:
		msg.Target.Contract = contract
	default:
		panic(fmt.Sprintf("unsupported contract type %T", contract))
	}

	return ch.CallView(msg)
}

// CallViewEx is a shortcut for CallView
func (ch *Chain) CallViewEx(c, ep string, params ...isc.CallArguments) (isc.CallArguments, error) {
	return ch.CallView(isc.NewMessageFromNames(c, ep, params...))
}

func (ch *Chain) CallViewAtState(chainState state.State, msg isc.Message) (isc.CallArguments, error) {
	return ch.callViewByHnameAtState(chainState, msg)
}

func (ch *Chain) CallViewByHname(msg isc.Message) (isc.CallArguments, error) {
	latestState, err := ch.store.LatestState()
	require.NoError(ch.Env.T, err)
	return ch.callViewByHnameAtState(latestState, msg)
}

func (ch *Chain) callViewByHnameAtState(chainState state.State, msg isc.Message) (isc.CallArguments, error) {
	ch.Log().Debugf("callView: %s::%s", msg.Target.Contract, msg.Target.EntryPoint)

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vmctx, err := viewcontext.New(
		ch.GetLatestAnchor(),
		chainState,
		ch.proc,
		ch.log,
		false,
	)
	if err != nil {
		return nil, err
	}
	return vmctx.CallViewExternal(msg)
}

// GetMerkleProofRaw returns Merkle proof of the key in the state
func (ch *Chain) GetMerkleProofRaw(key []byte) *trie.MerkleProof {
	ch.Log().Debugf("GetMerkleProof")

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	latestState, err := ch.LatestState()
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(
		ch.GetLatestAnchor(),
		latestState,
		ch.proc,
		ch.log,
		false,
	)
	require.NoError(ch.Env.T, err)
	ret, err := vmctx.GetMerkleProof(key)
	require.NoError(ch.Env.T, err)
	return ret
}

// GetBlockProof returns Merkle proof of the key in the state
func (ch *Chain) GetBlockProof(blockIndex uint32) (*blocklog.BlockInfo, *trie.MerkleProof, error) {
	ch.Log().Debugf("GetBlockProof")

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	latestState, err := ch.LatestState()
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(
		ch.GetLatestAnchor(),
		latestState,
		ch.proc,
		ch.log,
		false,
	)
	if err != nil {
		return nil, nil, err
	}
	retBlockInfo, retProof, err := vmctx.GetBlockProof(blockIndex)
	if err != nil {
		return nil, nil, err
	}
	return retBlockInfo, retProof, nil
}

// GetMerkleProof return the merkle proof of the key in the smart contract. Assumes Merkle model is used
func (ch *Chain) GetMerkleProof(scHname isc.Hname, key []byte) *trie.MerkleProof {
	return ch.GetMerkleProofRaw(append(scHname.Bytes(), key...))
}

// GetL1Commitment returns state commitment taken from the anchor output
func (ch *Chain) GetL1Commitment() *state.L1Commitment {
	anchorRef := ch.GetLatestAnchor()
	ret, err := transaction.L1CommitmentFromAnchor(anchorRef)
	require.NoError(ch.Env.T, err)
	return ret
}

// GetRootCommitment returns the root commitment of the latest state index
func (ch *Chain) GetRootCommitment() trie.Hash {
	block, err := ch.store.LatestBlock()
	require.NoError(ch.Env.T, err)
	return block.TrieRoot()
}

// GetContractStateCommitment returns commitment to the state of the specific contract, if possible
func (ch *Chain) GetContractStateCommitment(hn isc.Hname) ([]byte, error) {
	latestState, err := ch.LatestState()
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(
		ch.GetLatestAnchor(),
		latestState,
		ch.proc,
		ch.log,
		false,
	)
	if err != nil {
		return nil, err
	}
	return vmctx.GetContractStateCommitment(hn)
}

// WaitUntil waits until the condition specified by the given predicate yields true
func (env *Solo) WaitUntil(p func() bool, maxWait ...time.Duration) bool {
	env.T.Helper()
	maxw := 10 * time.Second
	var deadline time.Time
	if len(maxWait) > 0 {
		maxw = maxWait[0]
	}
	deadline = time.Now().Add(maxw)
	for {
		if p() {
			return true
		}
		if time.Now().After(deadline) {
			env.T.Logf("WaitUntil failed waiting max %v", maxw)
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

	deadline := time.Now().Add(realTimeout)
	for {
		if ch.mempool.Info().TotalPool == 0 {
			return true
		}
		time.Sleep(10 * time.Millisecond)
		if time.Now().After(deadline) {
			return false
		}
	}
}

// WaitForRequestsMark marks the amount of requests processed until now
// This allows the WaitForRequestsThrough() function to wait for the
// specified of number of requests after the mark point.
func (ch *Chain) WaitForRequestsMark() {
	ch.RequestsBlock = ch.LatestBlockIndex()
}

// WaitForRequestsThrough waits until the specified number of requests
// have been processed since the last call to WaitForRequestsMark()
func (ch *Chain) WaitForRequestsThrough(numReq int, maxWait ...time.Duration) bool {
	ch.Env.T.Helper()
	ch.Env.T.Logf("WaitForRequestsThrough: start -- block #%d -- numReq = %d", ch.RequestsBlock, numReq)
	return ch.Env.WaitUntil(func() bool {
		ch.Env.T.Helper()
		latest := ch.LatestBlockIndex()
		for ; ch.RequestsBlock < latest; ch.RequestsBlock++ {
			receipts := ch.GetRequestReceiptsForBlock(ch.RequestsBlock + 1)
			numReq -= len(receipts)
			ch.Env.T.Logf("WaitForRequestsThrough: new block #%d with %d requests -- numReq = %d", ch.RequestsBlock, len(receipts), numReq)
		}
		return numReq <= 0
	}, maxWait...)
}
