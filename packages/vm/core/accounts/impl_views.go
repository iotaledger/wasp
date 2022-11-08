package accounts

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
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
	return getAccountBalanceDict(getAccountR(ctx.StateR(), aid))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
// Params:
// - ParamAgentID
// Returns: {ParamBalance: uint64}
func viewBalanceBaseToken(ctx isc.SandboxView) dict.Dict {
	nTokens := getBaseTokensBalance(getAccountR(ctx.StateR(), ctx.Params().MustGetAgentID(ParamAgentID)))
	return dict.Dict{ParamBalance: codec.EncodeUint64(nTokens)}
}

// viewBalanceNativeToken returns the native token balance of the account belonging to the AgentID
// Params:
// - ParamAgentID
// - ParamNativeTokenID
// Returns: {ParamBalance: big.Int}
func viewBalanceNativeToken(ctx isc.SandboxView) dict.Dict {
	id := ctx.Params().MustGetNativeTokenID(ParamNativeTokenID)
	bal := getNativeTokenBalance(
		getAccountR(ctx.StateR(), ctx.Params().MustGetAgentID(ParamAgentID)),
		&id,
	)
	return dict.Dict{ParamBalance: bal.Bytes()}
}

// viewTotalAssets returns total balances controlled by the chain
func viewTotalAssets(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getAccountBalanceDict(getTotalL2AssetsAccountR(ctx.StateR()))
}

// viewAccounts returns list of all accounts as keys of the ImmutableCodec
func viewAccounts(ctx isc.SandboxView) dict.Dict {
	return getAccountsIntern(ctx.StateR())
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
	mapping := getNativeTokenOutputMapR(ctx.StateR())
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
	foundries := getAccountFoundriesR(ctx.StateR(), account)
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
	nftIDs := getAccountNFTs(getAccountR(ctx.StateR(), aid))

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamNFTIDs)
	for _, nftID := range nftIDs {
		arr.MustPush(nftID[:])
	}
	return ret
}

// viewNFTData returns the NFT data for a given NFTID
func viewNFTData(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewNFTData")
	nftIDBytes := ctx.Params().MustGetBytes(ParamNFTID)
	if len(nftIDBytes) != iotago.NFTIDLength {
		panic(ErrInvalidNFTID)
	}
	nftID := iotago.NFTID{}
	copy(nftID[:], nftIDBytes)
	data := GetNFTData(ctx.StateR(), nftID)
	return dict.Dict{
		ParamNFTData: data.Bytes(),
	}
}
