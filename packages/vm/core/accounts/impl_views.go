package accounts

import (
	"math"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// viewBalance returns the balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func viewBalance(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewBalance")
	aid, err := ctx.Params().GetAgentID(ParamAgentID)
	ctx.RequireNoError(err)
	return getAccountBalanceDict(ctx.StateR(), accountKey(aid))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
// Params:
// - ParamAgentID
// Returns: {ParamBalance: uint64}
func viewBalanceBaseToken(ctx isc.SandboxView) dict.Dict {
	nTokens := getBaseTokens(ctx.StateR(), accountKey(ctx.Params().MustGetAgentID(ParamAgentID)))
	return dict.Dict{ParamBalance: codec.EncodeUint64(nTokens)}
}

// viewBalanceNativeToken returns the native token balance of the account belonging to the AgentID
// Params:
// - ParamAgentID
// - ParamNativeTokenID
// Returns: {ParamBalance: big.Int}
func viewBalanceNativeToken(ctx isc.SandboxView) dict.Dict {
	nativeTokenID := ctx.Params().MustGetNativeTokenID(ParamNativeTokenID)
	bal := getNativeTokenAmount(
		ctx.StateR(),
		accountKey(ctx.Params().MustGetAgentID(ParamAgentID)),
		nativeTokenID,
	)
	return dict.Dict{ParamBalance: bal.Bytes()}
}

// viewTotalAssets returns total balances controlled by the chain
func viewTotalAssets(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getAccountBalanceDict(ctx.StateR(), l2TotalsAccount)
}

// viewAccounts returns list of all accounts
func viewAccounts(ctx isc.SandboxView) dict.Dict {
	return allAccountsAsDict(ctx.StateR())
}

// nonces are only sent with off-ledger requests
func viewGetAccountNonce(ctx isc.SandboxView) dict.Dict {
	account := ctx.Params().MustGetAgentID(ParamAgentID)
	nonce := GetMaxAssumedNonce(ctx.StateR(), account)
	ret := dict.New()
	ret.Set(ParamAccountNonce, codec.EncodeUint64(nonce))
	return ret
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chain
func viewGetNativeTokenIDRegistry(ctx isc.SandboxView) dict.Dict {
	mapping := nativeTokenOutputMapR(ctx.StateR())
	ret := dict.New()
	mapping.MustIterate(func(elemKey []byte, value []byte) bool {
		ret.Set(kv.Key(elemKey), []byte{0xff})
		return true
	})
	return ret
}

// viewAccountFoundries returns the foundries owned by the given agentID
func viewAccountFoundries(ctx isc.SandboxView) dict.Dict {
	account := ctx.Params().MustGetAgentID(ParamAgentID)
	foundries := accountFoundriesMapR(ctx.StateR(), account)
	ret := dict.New()
	foundries.MustIterate(func(k []byte, v []byte) bool {
		ret.Set(kv.Key(k), v)
		return true
	})
	return ret
}

// viewFoundryOutput takes serial number and returns corresponding foundry output in serialized form
func viewFoundryOutput(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewFoundryOutput")

	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	out, _, _ := GetFoundryOutput(ctx.StateR(), sn, ctx.ChainID())
	ctx.Requiref(out != nil, "foundry #%d does not exist", sn)
	outBin, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
	ctx.RequireNoError(err, "internal: error while serializing foundry output")
	ret := dict.New()
	ret.Set(ParamFoundryOutputBin, outBin)
	return ret
}

// viewAccountNFTs returns the NFTIDs of NFTs owned by an account
func viewAccountNFTs(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewAccountNFTs")
	aid := ctx.Params().MustGetAgentID(ParamAgentID)
	nftIDs := getAccountNFTs(ctx.StateR(), aid)

	if len(nftIDs) > math.MaxUint16 {
		panic("too many NFTs")
	}
	ret := dict.New()
	arr := collections.NewArray16(ret, ParamNFTIDs)
	for _, nftID := range nftIDs {
		nftID := nftID
		arr.MustPush(nftID[:])
	}
	return ret
}

func viewAccountNFTAmount(ctx isc.SandboxView) dict.Dict {
	aid := ctx.Params().MustGetAgentID(ParamAgentID)
	return dict.Dict{
		ParamNFTAmount: codec.EncodeUint32(nftsMapR(ctx.StateR(), aid).MustLen()),
	}
}

func viewAccountNFTsInCollection(ctx isc.SandboxView) dict.Dict {
	aid := ctx.Params().MustGetAgentID(ParamAgentID)
	collectionID := codec.MustDecodeNFTID(ctx.Params().MustGet(ParamCollectionID))
	nftIDs := getAccountNFTsInCollection(ctx.StateR(), aid, collectionID)

	if len(nftIDs) > math.MaxUint16 {
		panic("too many NFTs")
	}
	ret := dict.New()
	arr := collections.NewArray16(ret, ParamNFTIDs)
	for _, nftID := range nftIDs {
		nftID := nftID
		arr.MustPush(nftID[:])
	}
	return ret
}

func viewAccountNFTAmountInCollection(ctx isc.SandboxView) dict.Dict {
	aid := ctx.Params().MustGetAgentID(ParamAgentID)
	collectionID := codec.MustDecodeNFTID(ctx.Params().MustGet(ParamCollectionID))
	return dict.Dict{
		ParamNFTAmount: codec.EncodeUint32(nftsByCollectionMapR(ctx.StateR(), aid, kv.Key(collectionID[:])).MustLen()),
	}
}

// viewNFTData returns the NFT data for a given NFTID
func viewNFTData(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewNFTData")
	nftID := codec.MustDecodeNFTID(ctx.Params().MustGetBytes(ParamNFTID))
	data := MustGetNFTData(ctx.StateR(), nftID)
	return dict.Dict{
		ParamNFTData: data.Bytes(),
	}
}
