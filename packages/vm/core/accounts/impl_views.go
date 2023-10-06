package accounts

import (
	"math"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

// viewBalance returns the balances of the account belonging to the AgentID
// Params:
// - ParamAgentID (optional -- default: caller)
func viewBalance(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewBalance")
	aid := ctx.Params().MustGetAgentID(ParamAgentID, ctx.Caller())
	return getAccountBalanceDict(ctx.StateR(), accountKey(aid, ctx.ChainID()))
}

// viewBalanceBaseToken returns the base tokens balance of the account belonging to the AgentID
// Params:
// - ParamAgentID (optional -- default: caller)
func viewBalanceBaseToken(ctx isc.SandboxView) dict.Dict {
	nTokens := getBaseTokens(
		ctx.StateR(),
		accountKey(
			ctx.Params().MustGetAgentID(ParamAgentID, ctx.Caller()),
			ctx.ChainID(),
		),
	)
	return dict.Dict{ParamBalance: codec.EncodeUint64(nTokens)}
}

// viewBalanceNativeToken returns the native token balance of the account belonging to the AgentID
// Params:
// - ParamAgentID (optional -- default: caller)
// - ParamNativeTokenID
// Returns: {ParamBalance: big.Int}
func viewBalanceNativeToken(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	nativeTokenID := params.MustGetNativeTokenID(ParamNativeTokenID)
	bal := getNativeTokenAmount(
		ctx.StateR(),
		accountKey(params.MustGetAgentID(ParamAgentID, ctx.Caller()), ctx.ChainID()),
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
	account := ctx.Params().MustGetAgentID(ParamAgentID, ctx.Caller())
	nonce := AccountNonce(ctx.StateR(), account, ctx.ChainID())
	ret := dict.New()
	ret.Set(ParamAccountNonce, codec.EncodeUint64(nonce))
	return ret
}

// viewGetNativeTokenIDRegistry returns all native token ID accounted in the chain
func viewGetNativeTokenIDRegistry(ctx isc.SandboxView) dict.Dict {
	ret := dict.New()
	nativeTokenOutputMapR(ctx.StateR()).IterateKeys(func(tokenID []byte) bool {
		ret.Set(kv.Key(tokenID), []byte{0x01})
		return true
	})
	return ret
}

// viewAccountFoundries returns the foundries owned by the given agentID
func viewAccountFoundries(ctx isc.SandboxView) dict.Dict {
	ret := dict.New()
	account := ctx.Params().MustGetAgentID(ParamAgentID, ctx.Caller())
	accountFoundriesMapR(ctx.StateR(), account).IterateKeys(func(foundry []byte) bool {
		ret.Set(kv.Key(foundry), []byte{0x01})
		return true
	})
	return ret
}

var errFoundryNotFound = coreerrors.Register("foundry not found").Create()

// viewFoundryOutput takes serial number and returns corresponding foundry output in serialized form
func viewFoundryOutput(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewFoundryOutput")

	sn := ctx.Params().MustGetUint32(ParamFoundrySN)
	out, _ := GetFoundryOutput(ctx.StateR(), sn, ctx.ChainID())
	if out == nil {
		panic(errFoundryNotFound)
	}
	outBin, err := out.Serialize(serializer.DeSeriModeNoValidation, nil)
	ctx.RequireNoError(err, "internal: error while serializing foundry output")
	ret := dict.New()
	ret.Set(ParamFoundryOutputBin, outBin)
	return ret
}

// viewAccountNFTs returns the NFTIDs of NFTs owned by an account
func viewAccountNFTs(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewAccountNFTs")
	aid := ctx.Params().MustGetAgentID(ParamAgentID, ctx.Caller())
	nftIDs := getAccountNFTs(ctx.StateR(), aid)
	return listNFTIDs(nftIDs)
}

func viewAccountNFTAmount(ctx isc.SandboxView) dict.Dict {
	aid := ctx.Params().MustGetAgentID(ParamAgentID, ctx.Caller())
	return dict.Dict{
		ParamNFTAmount: codec.EncodeUint32(accountToNFTsMapR(ctx.StateR(), aid).Len()),
	}
}

func viewAccountNFTsInCollection(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	aid := params.MustGetAgentID(ParamAgentID, ctx.Caller())
	collectionID := params.MustGetNFTID(ParamCollectionID)
	nftIDs := getAccountNFTsInCollection(ctx.StateR(), aid, collectionID)
	return listNFTIDs(nftIDs)
}

func listNFTIDs(nftIDs []iotago.NFTID) dict.Dict {
	// TODO: add pagination?
	if len(nftIDs) > math.MaxUint16 {
		panic("too many NFTs")
	}
	ret := dict.New()
	arr := collections.NewArray(ret, ParamNFTIDs)
	for _, nftID := range nftIDs {
		nftID := nftID
		arr.Push(nftID[:])
	}
	return ret
}

func viewAccountNFTAmountInCollection(ctx isc.SandboxView) dict.Dict {
	params := ctx.Params()
	aid := params.MustGetAgentID(ParamAgentID, ctx.Caller())
	collectionID := params.MustGetNFTID(ParamCollectionID)
	return dict.Dict{
		ParamNFTAmount: codec.EncodeUint32(nftsByCollectionMapR(ctx.StateR(), aid, kv.Key(collectionID[:])).Len()),
	}
}

// viewNFTData returns the NFT data for a given NFTID
func viewNFTData(ctx isc.SandboxView) dict.Dict {
	ctx.Log().Debugf("accounts.viewNFTData")
	nftID := ctx.Params().MustGetNFTID(ParamNFTID)
	data := GetNFTData(ctx.StateR(), nftID)
	if data == nil {
		panic("NFTID not found")
	}
	return dict.Dict{
		ParamNFTData: data.Bytes(),
	}
}
