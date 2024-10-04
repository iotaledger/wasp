// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	vmerrors "github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
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
func NewCallParams(msg isc.Message) *CallParams {
	return &CallParams{msg: msg}
}

// NewCallParamsEx is a shortcut for NewCallParams
func NewCallParamsEx(c, ep string, params ...isc.CallArguments) *CallParams {
	return NewCallParams(isc.NewMessageFromNames(c, ep, params...))
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

func (r *CallParams) AddAllowanceNFTs(nftIDs ...sui.ObjectID) *CallParams {
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
func (r *CallParams) WithObject(objectID sui.ObjectID) *CallParams {
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

func (env *Solo) makeAssetsBag(keyPair *cryptolib.KeyPair, assets *isc.Assets) *sui.ObjectRef {
	ptb := sui.NewProgrammableTransactionBuilder()
	iscmoveclient.PTBAssetsBagNew(
		ptb,
		env.l1Config.ISCPackageID,
		keyPair.Address(),
	)
	assetsBagArg := ptb.LastCommandResultArg()

	allCoins := env.L1AllCoins(keyPair.Address())
	assets.Coins.IterateSorted(func(coinType coin.Type, amount coin.Value) bool {
		for _, ownedCoin := range allCoins {
			if ownedCoin.CoinType != coinType.String() {
				continue
			}
			coinRef := ownedCoin.Ref()
			coinArg := ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: coinRef})
			amountAdded := coin.Value(ownedCoin.Balance.Uint64())
			if amountAdded > amount {
				err := ptb.SplitCoin(coinRef, []uint64{uint64(amount)})
				require.NoError(env.T, err)
				coinArg = ptb.LastCommandResultArg()
				amountAdded = amount
			}
			iscmoveclient.PTBAssetsBagPlaceCoin(
				ptb,
				env.l1Config.ISCPackageID,
				assetsBagArg,
				coinArg,
				coinType.String(),
			)
			amount -= amountAdded
			if amount == 0 {
				break
			}
		}
		if amount > 0 {
			panic(fmt.Sprintf("makeAssetsBag: not enough L1 balance for coin %s", coinType))
		}
		return true
	})
	assets.Objects.IterateSorted(func(objectID sui.ObjectID) bool {
		panic("TODO")
	})
	res := env.executePTB(ptb.Finish(), keyPair)
	assetsBagRef, err := res.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(env.T, err)
	return assetsBagRef
}

// RequestFromParamsToLedger creates transaction with one request based on parameters and sigScheme
// Then it adds it to the ledger, atomically.
// Locking on the mutex is needed to prevent mess when several goroutines work on the same address
func (ch *Chain) RequestFromParamsToLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.RequestID, error) {
	res, err := ch.Env.ISCMoveClient().CreateAndSendRequest(
		ch.Env.ctx,
		keyPair,
		ch.Env.ISCPackageID(),
		ch.ID().AsAddress().AsSuiAddress(),
		ch.Env.makeAssetsBag(keyPair, req.assets),
		uint32(req.msg.Target.Contract),
		uint32(req.msg.Target.EntryPoint),
		req.msg.Params,
		nil, // Add allowance here
		req.gasBudget,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	if err != nil {
		return isc.RequestID{}, err
	}
	reqRef, err := res.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	if err != nil {
		return isc.RequestID{}, err
	}

	return isc.RequestID(*reqRef.ObjectID), nil
}

// PostRequestSync posts a request synchronously sent by the test program to the smart contract on the same or another chain:
//   - creates a request transaction with the request block on it. The sigScheme is used to
//     sign the inputs of the transaction or OriginatorKeyPair is used if parameter is nil
//   - adds request transaction to UTXODB
//   - runs the request in the VM. It results in new updated virtual state and a new transaction
//     which anchors the state.
//   - adds the resulting transaction to UTXODB
//   - posts requests, contained in the resulting transaction to backlog queues of respective chains
//   - returns the result of the call to the smart contract's entry point
//
// Note that in real network of Wasp nodes (the committee) posting the transaction is completely
// asynchronous, i.e. result of the call is not available to the originator of the post.
//
// Unlike the real Wasp environment, the 'solo' environment makes PostRequestSync a synchronous call.
// It makes it possible step-by-step debug of the smart contract logic.
// The call should be used only from the main thread (goroutine)
func (ch *Chain) PostRequestSync(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, error) {
	_, ret, err := ch.PostRequestSyncTx(req, keyPair)
	return ret, err
}

func (ch *Chain) PostRequestOffLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, error) {
	r := req.NewRequestOffLedger(ch, keyPair)
	return ch.RunOffLedgerRequest(r)
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, keyPair *cryptolib.KeyPair) (isc.RequestID, isc.CallArguments, error) {
	reqID, receipt, res, err := ch.PostRequestSyncExt(req, keyPair)
	if err != nil {
		return reqID, res, err
	}
	return reqID, res, ch.ResolveVMError(receipt.Error).AsGoError()
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

func (ch *Chain) PostRequestSyncExt(callParams *CallParams, keyPair *cryptolib.KeyPair) (isc.RequestID, *blocklog.RequestReceipt, isc.CallArguments, error) {
	defer ch.logRequestLastBlock()

	reqID, err := ch.RequestFromParamsToLedger(callParams, keyPair)
	require.NoError(ch.Env.T, err)
	reqWithObj, err := ch.Env.ISCMoveClient().GetRequestFromObjectID(ch.Env.ctx, (*sui.ObjectID)(&reqID))
	req, err := isc.OnLedgerFromRequest(reqWithObj, keyPair.Address())
	results := ch.RunRequestsSync([]isc.Request{req}, "post")
	if len(results) == 0 {
		return isc.RequestID{}, nil, nil, errors.New("request has been skipped")
	}
	res := results[0]
	return reqID, res.Receipt, res.Return, nil
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
	latestState, err := ch.LatestState(chain.ActiveOrCommittedState)
	if err != nil {
		return nil, err
	}
	return ch.CallViewAtState(latestState, msg)
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

	vmctx, err := viewcontext.New(ch, chainState, false)
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

	latestState, err := ch.LatestState(chain.ActiveOrCommittedState)
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(ch, latestState, false)
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

	latestState, err := ch.LatestState(chain.ActiveOrCommittedState)
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(ch, latestState, false)
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
	ret, err := transaction.L1CommitmentFromAnchor(anchorRef.Object)
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
	latestState, err := ch.LatestState(chain.ActiveOrCommittedState)
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(ch, latestState, false)
	if err != nil {
		return nil, err
	}
	return vmctx.GetContractStateCommitment(hn)
}

// WaitUntil waits until the condition specified by the given predicate yields true
func (ch *Chain) WaitUntil(p func() bool, maxWait ...time.Duration) bool {
	ch.Env.T.Helper()
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
			ch.Env.T.Logf("WaitUntil failed waiting max %v", maxw)
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
	return ch.WaitUntil(func() bool {
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
