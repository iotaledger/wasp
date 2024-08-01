package accounts

import (
	"math/big"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// viewBalance returns the balances of the account belonging to the AgentID
func viewBalance(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *isc.Assets {
	ctx.Log().Debugf("accounts.viewBalance")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getFungibleTokens(accountKey(aid, ctx.ChainID()))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
func viewBalanceBaseToken(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).GetBaseTokensBalanceDiscardExtraDecimals(aid, ctx.ChainID())
}

// viewBalanceBaseTokenEVM returns the base tokens balance of the account belonging to the AgentID (in the EVM format with 18 decimals)
func viewBalanceBaseTokenEVM(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).GetBaseTokensBalanceFullDecimals(aid, ctx.ChainID())
}

// viewBalanceNativeToken returns the native token balance of the account belonging to the AgentID
func viewBalanceNativeToken(ctx isc.SandboxView, optionalAgentID *isc.AgentID, nativeTokenID isc.CoinType) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getNativeTokenAmount(
		accountKey(aid, ctx.ChainID()),
		nativeTokenID,
	)
}

// viewTotalAssets returns total balances controlled by the chain
func viewTotalAssets(ctx isc.SandboxView) *isc.Assets {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return NewStateReaderFromSandbox(ctx).getFungibleTokens(L2TotalsAccount)
}

// nonces are only sent with off-ledger requests
func viewGetAccountNonce(ctx isc.SandboxView, optionalAgentID *isc.AgentID) uint64 {
	account := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).AccountNonce(account, ctx.ChainID())
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chain
func viewGetNativeTokenIDRegistry(ctx isc.SandboxView) []isc.CoinType {
	ntMap := NewStateReaderFromSandbox(ctx).nativeTokenOutputMapR()
	ret := make([]isc.CoinType, 0, ntMap.Len())
	ntMap.IterateKeys(func(b []byte) bool {
		ntID := codec.CoinType.MustDecode(b)
		ret = append(ret, ntID)
		return true
	})
	return ret
}

// viewAccountFoundries returns the foundries owned by the given agentID
func viewAccountFoundries(ctx isc.SandboxView, optionalAgentID *isc.AgentID) map[uint32]struct{} {
	ret := make(map[uint32]struct{})
	account := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	NewStateReaderFromSandbox(ctx).accountFoundriesMapR(account).IterateKeys(func(b []byte) bool {
		ret[codec.Uint32.MustDecode(b)] = struct{}{}
		return true
	})
	return ret
}

var errFoundryNotFound = coreerrors.Register("foundry not found").Create()

// viewFoundryOutput takes serial number and returns corresponding foundry output in serialized form
func viewFoundryOutput(ctx isc.SandboxView, sn uint32) sui.ObjectID {
	ctx.Log().Debugf("accounts.viewFoundryOutput")
	out, _ := NewStateReaderFromSandbox(ctx).GetFoundryOutput(sn, ctx.ChainID())
	if out == nil {
		panic(errFoundryNotFound)
	}
	// TODO: refactor me (FoundryOutput -> TreasuryCap -> Sui.ObjectID)
	return out
}

// viewAccountNFTs returns the NFTIDs of NFTs owned by an account
func viewAccountNFTs(ctx isc.SandboxView, optionalAgentID *isc.AgentID) []sui.ObjectID {
	ctx.Log().Debugf("accounts.viewAccountNFTs")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getAccountNFTs(aid)
}

func viewAccountNFTAmount(ctx isc.SandboxView, optionalAgentID *isc.AgentID) uint32 {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).accountToNFTsMapR(aid).Len()
}

func viewAccountNFTsInCollection(ctx isc.SandboxView, optionalAgentID *isc.AgentID, collectionID sui.ObjectID) []sui.ObjectID {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).getAccountNFTsInCollection(aid, collectionID)
}

func viewAccountNFTAmountInCollection(ctx isc.SandboxView, optionalAgentID *isc.AgentID, collectionID sui.ObjectID) uint32 {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return NewStateReaderFromSandbox(ctx).nftsByCollectionMapR(aid, kv.Key(collectionID[:])).Len()
}

// viewNFTData returns the NFT data for a given NFTID
func viewNFTData(ctx isc.SandboxView, nftID sui.ObjectID) *isc.NFT {
	ctx.Log().Debugf("accounts.viewNFTData")
	data := NewStateReaderFromSandbox(ctx).GetNFTData(nftID)
	if data == nil {
		panic("NFTID not found")
	}
	return data
}
