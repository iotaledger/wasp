package accounts

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// viewBalance returns the balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func viewBalance(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewBalance")
	aid, err := ctx.Params().GetAgentID(ParamAgentID)
	ctx.RequireNoError(err)
	return getAccountBalanceDict(getAccountR(ctx.State(), aid))
}

// viewBalanceIotas returns the iota balance of the account belonging to the AgentID
// Params:
// - ParamAgentID
// Returns: {ParamBalance: uint64}
func viewBalanceIotas(ctx iscp.SandboxView) dict.Dict {
	iotas := getIotaBalance(getAccountR(ctx.State(), ctx.Params().MustGetAgentID(ParamAgentID)))
	return dict.Dict{ParamBalance: codec.EncodeUint64(iotas)}
}

// viewBalanceNativeToken returns the native token balance of the account belonging to the AgentID
// Params:
// - ParamAgentID
// - ParamNativeTokenID
// Returns: {ParamBalance: big.Int}
func viewBalanceNativeToken(ctx iscp.SandboxView) dict.Dict {
	id := ctx.Params().MustGetNativeTokenID(ParamNativeTokenID)
	bal := getNativeTokenBalance(
		getAccountR(ctx.State(), ctx.Params().MustGetAgentID(ParamAgentID)),
		&id,
	)
	return dict.Dict{ParamBalance: bal.Bytes()}
}

// viewTotalAssets returns total colored balances controlled by the chain
func viewTotalAssets(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewTotalAssets")
	return getAccountBalanceDict(getTotalL2AssetsAccountR(ctx.State()))
}

// viewAccounts returns list of all accounts as keys of the ImmutableCodec
func viewAccounts(ctx iscp.SandboxView) dict.Dict {
	return getAccountsIntern(ctx.State())
}

// nonces are only sent with off-ledger requests
func viewGetAccountNonce(ctx iscp.SandboxView) dict.Dict {
	account := ctx.Params().MustGetAgentID(ParamAgentID)
	nonce := GetMaxAssumedNonce(ctx.State(), account)
	ret := dict.New()
	ret.Set(ParamAccountNonce, codec.EncodeUint64(nonce))
	return ret
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chian
func viewGetNativeTokenIDRegistry(ctx iscp.SandboxView) dict.Dict {
	mapping := getNativeTokenOutputMapR(ctx.State())
	ret := dict.New()
	mapping.MustIterate(func(elemKey []byte, value []byte) bool {
		ret.Set(kv.Key(elemKey), []byte{0xff})
		return true
	})
	return ret
}

// viewFoundryOutput takes serial number and returns corresponding foundry output in serialized form
func viewFoundryOutput(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewFoundryOutput")

	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	out, _, _ := GetFoundryOutput(ctx.State(), sn, ctx.ChainID())
	ctx.Requiref(out != nil, "foundry #%d does not exist", sn)
	outBin, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
	ctx.RequireNoError(err, "internal: error while serializing foundry output")
	ret := dict.New()
	ret.Set(ParamFoundryOutputBin, outBin)
	return ret
}

// viewAccountNFTs returns the NFTIDs of NFTs owned by an account
func viewAccountNFTs(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewAccountNFTs")
	aid := ctx.Params().MustGetAgentID(ParamAgentID)
	nftIDs := getAccountNFTs(getAccountR(ctx.State(), aid))

	ret := dict.New()
	arr := collections.NewArray16(ret, ParamNFTIDs)
	for _, nftID := range nftIDs {
		arr.MustPush(nftID[:])
	}
	return ret
}

// viewNFTData returns the NFT data for a given NFTID
func viewNFTData(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewNFTData")
	nftIDBytes := ctx.Params().MustGetBytes(ParamNFTID)
	if len(nftIDBytes) != iotago.NFTIDLength {
		panic(ErrInvalidNFTID)
	}
	nftID := iotago.NFTID{}
	copy(nftID[:], nftIDBytes)
	data := GetNFTData(ctx.State(), nftID)
	return dict.Dict{
		ParamNFTData: data.Bytes(),
	}
}
