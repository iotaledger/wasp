// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"errors"
	"fmt"
	"math"

	"fortio.org/safecast"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	vmerrors "github.com/iotaledger/wasp/v2/packages/vm/core/errors"
	"github.com/iotaledger/wasp/v2/packages/vm/viewcontext"
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

func (r *CallParams) AddAllowanceCoins(coinType coin.Type, amount coin.Value) *CallParams {
	if r.allowance == nil {
		r.allowance = isc.NewEmptyAssets()
	}

	r.allowance.AddCoin(coinType, amount)

	return r
}

func (r *CallParams) AddAllowanceObject(obj isc.IotaObject) *CallParams {
	if r.allowance == nil {
		r.allowance = isc.NewEmptyAssets()
	}
	r.allowance.AddObject(obj)
	return r
}

// WithAssets sets the assets to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) WithAssets(assets *isc.Assets) *CallParams {
	r.assets = assets.Clone()
	return r
}

// WithFungibleTokens sets the tokens to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) WithFungibleTokens(ftokens isc.CoinBalances) *CallParams {
	r.assets.Coins = ftokens.Clone()
	return r
}

// AddFungibleTokens adds tokens to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) AddFungibleTokens(ftokens isc.CoinBalances) *CallParams {
	r.assets.Add(ftokens.ToAssets())
	return r
}

// AddBaseTokens adds base tokens to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) AddBaseTokens(amount coin.Value) *CallParams {
	r.assets.AddBaseTokens(amount)
	return r
}

// AddCoin adds a coin to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) AddCoin(coinType coin.Type, amount coin.Value) *CallParams {
	r.assets.AddCoin(coinType, amount)
	return r
}

// AddObject adds an object to be sent (only applicable when the call is made via on-ledger request)
func (r *CallParams) AddObject(obj isc.IotaObject) *CallParams {
	r.assets.AddObject(obj)
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

// NewRequestOnLedger creates an on-ledger request without sending it to L1. It is intended
// mainly for estimating gas.
func (r *CallParams) NewRequestOnLedger(ch *Chain, keyPair *cryptolib.KeyPair) (isc.OnLedgerRequest, error) {
	if keyPair == nil {
		keyPair = ch.ChainAdmin
	}
	ref := iotatest.RandomObjectRef()
	assetsBagRef := iotatest.RandomObjectRef()
	return isc.OnLedgerFromMoveRequest(&iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *ref,
		Object: &iscmove.Request{
			ID:     *ref.ObjectID,
			Sender: keyPair.Address(),
			AssetsBag: *r.assets.AsAssetsBagWithBalances(&iscmove.AssetsBag{
				ID:   *assetsBagRef.ObjectID,
				Size: safecast.MustConvert[uint64](r.assets.Size()),
			}),
			Message: iscmove.Message{
				Contract: uint32(r.msg.Target.Contract),
				Function: uint32(r.msg.Target.EntryPoint),
				Args:     r.msg.Params,
			},
			AllowanceBCS: lo.Must(bcs.Marshal(r.allowance.AsISCMove())),
			GasBudget:    r.gasBudget,
		},
	}, ch.ChainID.AsAddress())
}

// NewRequestOffLedger creates off-ledger request from parameters
func (r *CallParams) NewRequestOffLedger(ch *Chain, keyPair *cryptolib.KeyPair) isc.OffLedgerRequest {
	if keyPair == nil {
		keyPair = ch.ChainAdmin
	}
	if r.nonce == 0 {
		r.nonce = ch.Nonce(isc.NewAddressAgentID(keyPair.Address()))
	}
	ret := isc.NewOffLedgerRequest(ch.ChainID, r.msg, r.nonce, r.gasBudget).
		WithAllowance(r.allowance)
	return ret.Sign(keyPair)
}

func (env *Solo) SelectCoinsForGas(
	addr *cryptolib.Address,
	targetPTB *iotago.ProgrammableTransaction,
	gasBudget uint64,
) []*iotago.ObjectRef {
	pickedCoins, err := iotagraphql.PickupCoinsWithFilter(
		env.L1BaseTokenCoins(addr),
		gasBudget,
		func(c iotagraphql.Coin) bool {
			id := c.ObjectID()
			return !targetPTB.IsInInputObjects(&id)
		},
	)
	require.NoError(env.T, err)
	refs, err := pickedCoins.CoinRefs()
	require.NoError(env.T, err)
	return refs
}

func (env *Solo) makeBaseTokenCoin(
	keyPair *cryptolib.KeyPair,
	value coin.Value,
	filter func(iotagraphql.Coin) bool,
) *iotago.ObjectRef {
	allCoins := env.L1BaseTokenCoins(keyPair.Address())
	require.NotEmpty(env.T, allCoins)

	const gasBudget = iotagraphql.DefaultGasBudget

	pickedCoin, ok, err := iotagraphql.PickupCoinWithFilter(
		env.L1BaseTokenCoins(keyPair.Address()),
		uint64(value+gasBudget),
		filter,
	)

	require.NoError(env.T, err)
	require.True(env.T, ok, "no coin found with sufficient balance")

	pickedCoinID := pickedCoin.ObjectID()
	tx := lo.Must(env.L1Client().PayIota(
		env.ctx,
		iotagraphql.PayIotaRequest{
			Signer:     keyPair.Address().AsIotaAddress(),
			InputCoins: []iotago.ObjectID{pickedCoinID},
			Amount:     []*iotagraphql.BigInt{iotagraphql.NewBigInt(uint64(value))},
			Recipients: []*iotago.Address{lo.ToPtr(keyPair.Address().AsIotaAddress())},
			GasBudget:  iotagraphql.NewBigInt(gasBudget),
		},
	))

	pickedCoinRef, err := pickedCoin.ObjectRef()
	require.NoError(env.T, err)
	var baseTokenCoin *iotago.ObjectRef = nil
	env.MustWithWaitForNextVersion(pickedCoinRef, func() {
		txnResponse, err := env.L1Client().SignAndExecuteTransaction(
			env.ctx,
			tx.TxBytes,
			cryptolib.SignerToIotaSigner(keyPair),
		)

		require.NoError(env.T, err)
		require.True(env.T, txnResponse.ExecuteTransactionBlock.Effects.IsSuccess())

		changes := txnResponse.ExecuteTransactionBlock.Effects.GetObjectChanges().Nodes
		for i := range changes {
			if changes[i].GetIdCreated() {
				ref, refErr := changes[i].OutputState.ObjectRef()
				require.NoError(env.T, refErr)
				baseTokenCoin = ref
			}
		}
		require.NotNil(env.T, baseTokenCoin)
	})

	return baseTokenCoin
}

func (ch *Chain) SendRequestWithL1GasBudget(
	req *CallParams,
	keyPair *cryptolib.KeyPair,
	l1GasBudget uint64,
) (
	isc.OnLedgerRequest,
	*iotagraphql.ExecuteTransactionBlockResponse,
	error,
) {
	if keyPair == nil {
		keyPair = ch.ChainAdmin
	}
	res, err := ch.Env.ISCMoveClient().CreateAndSendRequestWithAssets(
		ch.Env.ctx,
		&iscmoveclient.CreateAndSendRequestWithAssetsRequest{
			Signer:        keyPair,
			PackageID:     ch.Env.ISCPackageID(),
			AnchorAddress: lo.ToPtr(ch.ID().AsAddress().AsIotaAddress()),
			Assets:        req.assets.AsISCMove(),
			Message: &iscmove.Message{
				Contract: uint32(req.msg.Target.Contract),
				Function: uint32(req.msg.Target.EntryPoint),
				Args:     req.msg.Params,
			},
			AllowanceBCS:     lo.Must(bcs.Marshal(req.allowance.AsISCMove())),
			OnchainGasBudget: req.gasBudget,
			GasPrice:         iotagraphql.DefaultGasPrice,
			GasBudget:        l1GasBudget,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	reqRef, err := res.GetCreatedObjectByName(iscmove.RequestModuleName, iscmove.RequestObjectName)
	require.NoError(ch.Env.T, err)

	l1req := ch.GetL1RequestData(*reqRef.ObjectID)
	return l1req, res, nil
}

func (ch *Chain) GetL1RequestData(objectID iotago.ObjectID) isc.OnLedgerRequest {
	reqWithObj, err := ch.Env.ISCMoveClient().GetRequestFromObjectID(ch.Env.ctx, &objectID)
	require.NoError(ch.Env.T, err)

	r, err := isc.OnLedgerFromMoveRequest(reqWithObj, ch.ID().AsAddress())
	require.NoError(ch.Env.T, err)
	return r
}

// SendRequest creates a request based on parameters and sigScheme, then send it to the anchor.
func (ch *Chain) SendRequest(req *CallParams, keyPair *cryptolib.KeyPair) (isc.OnLedgerRequest, *iotagraphql.ExecuteTransactionBlockResponse, error) {
	return ch.SendRequestWithL1GasBudget(req, keyPair, iotagraphql.DefaultGasBudget)
}

// PostRequestSync posts a request synchronously sent by the test program to
// the smart contract on the same or another chain.
//
// Note that in a real network of Wasp nodes, posting a transaction is completely
// asynchronous, i.e. result of the call is not available to the originator of the post.
// Instead, the Solo environment makes PostRequestSync a synchronous call,
// making it possible to step-by-step debug the smart contract logic.
func (ch *Chain) PostRequestSync(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, error) {
	_, _, res, _, err := ch.PostRequestSyncTx(req, keyPair)
	if err != nil {
		return nil, err
	}
	return res.Return, nil
}

func (ch *Chain) PostRequestOffLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, error) {
	r := req.NewRequestOffLedger(ch, keyPair)
	_, res, err := ch.RunOffLedgerRequest(r)
	return res, err
}

func (ch *Chain) PostRequestSyncTx(req *CallParams, keyPair *cryptolib.KeyPair) (
	onLedregReq isc.OnLedgerRequest,
	l1Res *iotagraphql.ExecuteTransactionBlockResponse,
	vmRes *vm.RequestResult,
	anchorTransitionPTBRes *iotagraphql.ExecuteTransactionBlockResponse,
	err error,
) {
	onLedregReq, l1Res, vmRes, anchorTransitionPTBRes, err = ch.PostRequestSyncExt(req, keyPair)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return onLedregReq, l1Res, vmRes, anchorTransitionPTBRes, ch.ResolveVMError(vmRes.Receipt.Error).AsGoError()
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
	req isc.OnLedgerRequest,
	l1Res *iotagraphql.ExecuteTransactionBlockResponse,
	vmRes *vm.RequestResult,
	anchorTransitionPTBRes *iotagraphql.ExecuteTransactionBlockResponse,
	err error,
) {
	if keyPair == nil {
		keyPair = ch.ChainAdmin
	}
	defer ch.logRequestLastBlock()

	req, l1Res, err = ch.SendRequest(callParams, keyPair)
	require.NoError(ch.Env.T, err, "failed to send the request")

	anchorTransitionPTBRes, results := ch.RunRequestsSync([]isc.Request{req})
	if len(results) == 0 {
		return nil, nil, nil, nil, errors.New("request has been skipped")
	}
	require.Len(ch.Env.T, results, 1)
	return req, l1Res, results[0], anchorTransitionPTBRes, nil
}

// EstimateGasOnLedger executes the given on-ledger request without committing
// any changes in the ledger. It returns the amount of gas consumed.
// WARNING: Gas estimation is just an "estimate", there is no guarantees that the real call will bear the same cost, due to the turing-completeness of smart contracts
// TODO only a senderAddr, not a keyPair should be necessary to estimate (it definitely shouldn't fallback to the chain originator)
func (ch *Chain) EstimateGasOnLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, *blocklog.RequestReceipt, error) {
	r, err := req.NewRequestOnLedger(ch, keyPair)
	if err != nil {
		return nil, nil, err
	}
	res := ch.EstimateGas(r)
	return res.Return, res.Receipt, ch.ResolveVMError(res.Receipt.Error).AsGoError()
}

// EstimateGasOffLedger executes the given on-ledger request without committing
// any changes in the ledger. It returns the amount of gas consumed.
// WARNING: Gas estimation is just an "estimate", there is no guarantees that the real call will bear the same cost, due to the turing-completeness of smart contracts
func (ch *Chain) EstimateGasOffLedger(req *CallParams, keyPair *cryptolib.KeyPair) (isc.CallArguments, *blocklog.RequestReceipt, error) {
	reqCopy := *req
	r := reqCopy.NewRequestOffLedger(ch, keyPair)
	res := ch.EstimateGas(r)
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
	ch.Log().LogDebugf("callView: %s::%s", msg.Target.Contract, msg.Target.EntryPoint)

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	vmctx, err := viewcontext.New(
		ch.ChainID,
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
	ch.Log().LogDebugf("GetMerkleProof")

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	latestState, err := ch.LatestState()
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(
		ch.ChainID,
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
	ch.Log().LogDebugf("GetBlockProof")

	ch.runVMMutex.Lock()
	defer ch.runVMMutex.Unlock()

	latestState, err := ch.LatestState()
	require.NoError(ch.Env.T, err)
	vmctx, err := viewcontext.New(
		ch.ChainID,
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
		ch.ChainID,
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
