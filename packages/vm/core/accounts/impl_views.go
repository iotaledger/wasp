package accounts

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// viewBalance returns colored balances of the account belonging to the AgentID
// Params:
// - ParamAgentID
func viewBalance(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewBalance")
	aid, err := ctx.Params().GetAgentID(ParamAgentID)
	ctx.RequireNoError(err)
	return getAccountBalanceDict(getAccountR(ctx.State(), aid))
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

func viewGetAccountNonce(ctx iscp.SandboxView) dict.Dict {
	account := ctx.Params().MustGetAgentID(ParamAgentID)
	nonce := GetMaxAssumedNonce(ctx.State(), account.Address())
	ret := dict.New()
	ret.Set(ParamAccountNonce, codec.EncodeUint64(nonce))
	return ret
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chian
func viewGetNativeTokenIDRegistry(ctx iscp.SandboxView) dict.Dict {
	mapping := getNativeTokenOutputMapR(ctx.State())
	ret := dict.New()
	mapping.MustIterate(func(elemKey []byte, value []byte) bool {
		ret.Set(kv.Key(elemKey), []byte{0xFF})
		return true
	})
	return ret
}

// viewAccountNFTs returns the NFTIDs of NFTs owned by an account
func viewAccountNFTs(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewAccountNFTs")
	aid := ctx.Params().MustGetAgentID(ParamAgentID)
	nftIDs := getAccountNFTs(getAccountR(ctx.State(), aid))
	ret := []byte{}
	for _, nftid := range nftIDs {
		ret = append(ret, nftid[:]...)
	}
	return dict.Dict{
		ParamNFTIDs: ret,
	}
}

// viewFoundryOutput takes serial number and returns corresponding foundry output in serialized form
func viewFoundryOutput(ctx iscp.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.foundryOutput")

	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	out, _, _ := GetFoundryOutput(ctx.State(), sn, ctx.ChainID())
	ctx.Requiref(out != nil, "foundry #%d does not exist", sn)
	outBin, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
	ctx.RequireNoError(err, "internal: error while serializing foundry output")
	ret := dict.New()
	ret.Set(ParamFoundryOutputBin, outBin)
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
