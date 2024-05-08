package accounts

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

// viewBalance returns the balances of the account belonging to the AgentID
func viewBalance(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *isc.Assets {
	ctx.Log().Debugf("accounts.viewBalance")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return getFungibleTokens(ctx.SchemaVersion(), ctx.StateR(), accountKey(aid, ctx.ChainID()))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
func viewBalanceBaseToken(ctx isc.SandboxView, optionalAgentID *isc.AgentID) uint64 {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return getBaseTokens(ctx.SchemaVersion())(
		ctx.StateR(),
		accountKey(aid, ctx.ChainID()),
	)
}

// viewBalanceBaseTokenEVM returns the base tokens balance of the account belonging to the AgentID (in the EVM format with 18 decimals)
func viewBalanceBaseTokenEVM(ctx isc.SandboxView, optionalAgentID *isc.AgentID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return GetBaseTokensFullDecimals(ctx.SchemaVersion())(
		ctx.StateR(),
		accountKey(aid, ctx.ChainID()),
	)
}

// viewBalanceNativeToken returns the native token balance of the account belonging to the AgentID
func viewBalanceNativeToken(ctx isc.SandboxView, optionalAgentID *isc.AgentID, nativeTokenID iotago.NativeTokenID) *big.Int {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return getNativeTokenAmount(
		ctx.StateR(),
		accountKey(aid, ctx.ChainID()),
		nativeTokenID,
	)
}

// viewTotalAssets returns total balances controlled by the chain
func viewTotalAssets(ctx isc.SandboxView) *isc.Assets {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getFungibleTokens(ctx.SchemaVersion(), ctx.StateR(), L2TotalsAccount)
}

// viewAccounts returns list of all accounts
func viewAccounts(ctx isc.SandboxView) dict.Dict {
	return AllAccountsAsDict(ctx.StateR())
}

// nonces are only sent with off-ledger requests
func viewGetAccountNonce(ctx isc.SandboxView, optionalAgentID *isc.AgentID) uint64 {
	account := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return AccountNonce(ctx.StateR(), account, ctx.ChainID())
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chain
func viewGetNativeTokenIDRegistry(ctx isc.SandboxView) []iotago.NativeTokenID {
	ntMap := nativeTokenOutputMapR(ctx.StateR())
	ret := make([]iotago.NativeTokenID, 0, ntMap.Len())
	ntMap.IterateKeys(func(b []byte) bool {
		ntID := codec.NativeTokenID.MustDecode(b)
		ret = append(ret, ntID)
		return true
	})
	return ret
}

// viewAccountFoundries returns the foundries owned by the given agentID
func viewAccountFoundries(ctx isc.SandboxView, optionalAgentID *isc.AgentID) map[uint32]struct{} {
	ret := make(map[uint32]struct{})
	account := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	accountFoundriesMapR(ctx.StateR(), account).IterateKeys(func(b []byte) bool {
		ret[codec.Uint32.MustDecode(b)] = struct{}{}
		return true
	})
	return ret
}

var errFoundryNotFound = coreerrors.Register("foundry not found").Create()

// viewFoundryOutput takes serial number and returns corresponding foundry output in serialized form
func viewFoundryOutput(ctx isc.SandboxView, sn uint32) iotago.Output {
	ctx.Log().Debugf("accounts.viewFoundryOutput")
	out, _ := GetFoundryOutput(ctx.StateR(), sn, ctx.ChainID())
	if out == nil {
		panic(errFoundryNotFound)
	}
	return out
}

// viewAccountNFTs returns the NFTIDs of NFTs owned by an account
func viewAccountNFTs(ctx isc.SandboxView, optionalAgentID *isc.AgentID) []iotago.NFTID {
	ctx.Log().Debugf("accounts.viewAccountNFTs")
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return getAccountNFTs(ctx.StateR(), aid)
}

func viewAccountNFTAmount(ctx isc.SandboxView, optionalAgentID *isc.AgentID) uint32 {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return accountToNFTsMapR(ctx.StateR(), aid).Len()
}

func viewAccountNFTsInCollection(ctx isc.SandboxView, optionalAgentID *isc.AgentID, collectionID iotago.NFTID) []iotago.NFTID {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return getAccountNFTsInCollection(ctx.StateR(), aid, collectionID)
}

func viewAccountNFTAmountInCollection(ctx isc.SandboxView, optionalAgentID *isc.AgentID, collectionID iotago.NFTID) uint32 {
	aid := coreutil.FromOptional(optionalAgentID, ctx.Caller())
	return nftsByCollectionMapR(ctx.StateR(), aid, kv.Key(collectionID[:])).Len()
}

// viewNFTData returns the NFT data for a given NFTID
func viewNFTData(ctx isc.SandboxView, nftID iotago.NFTID) *isc.NFT {
	ctx.Log().Debugf("accounts.viewNFTData")
	data := GetNFTData(ctx.StateR(), nftID)
	if data == nil {
		panic("NFTID not found")
	}
	return data
}
