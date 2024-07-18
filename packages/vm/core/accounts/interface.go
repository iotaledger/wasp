package accounts

import (
	"math"
	"math/big"

	"github.com/samber/lo"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Contract = coreutil.NewContract(coreutil.CoreContractAccounts)

var (
	// Funcs
	FuncDeposit = coreutil.NewEP0(Contract, "deposit")
	// Kept for compatibility reasons
	FuncFoundryCreateNew = coreutil.NewEP1(Contract, "foundryCreateNew",
		coreutil.FieldWithCodecOptional(ParamTokenScheme, codec.TokenScheme),
	)
	// TODO implement grant/claim protocol of moving ownership of the foundry
	//  Including ownership of the foundry by the common account/chain owner
	FuncNativeTokenCreate       = EPNativeTokenCreate{EntryPointInfo: Contract.Func("nativeTokenCreate")}
	FuncNativeTokenModifySupply = EPNativeTokenModifySupply{EntryPointInfo: Contract.Func("nativeTokenModifySupply")}
	FuncNativeTokenDestroy      = coreutil.NewEP1(Contract, "nativeTokenDestroy",
		coreutil.FieldWithCodec(ParamFoundrySN, codec.Uint32),
	)
	FuncMintNFT                = EPMintNFT{EntryPointInfo: Contract.Func("mintNFT")}
	FuncTransferAccountToChain = coreutil.NewEP1(Contract, "transferAccountToChain",
		coreutil.FieldWithCodecOptional(ParamGasReserve, codec.Uint64),
	)
	FuncTransferAllowanceTo = coreutil.NewEP1(Contract, "transferAllowanceTo",
		coreutil.FieldWithCodec(ParamAgentID, codec.AgentID),
	)
	FuncWithdraw = coreutil.NewEP0(Contract, "withdraw")

	// Views
	ViewAccountFoundries = coreutil.NewViewEP11(Contract, "accountFoundries",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		OutputSerialNumberSet{},
	)
	ViewAccountNFTAmount = coreutil.NewViewEP11(Contract, "accountNFTAmount",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamNFTAmount, codec.Uint32),
	)
	ViewAccountNFTAmountInCollection = coreutil.NewViewEP21(Contract, "accountNFTAmountInCollection",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamCollectionID, codec.NFTID),
		coreutil.FieldWithCodec(ParamNFTAmount, codec.Uint32),
	)
	ViewAccountNFTs = coreutil.NewViewEP11(Contract, "accountNFTs",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		OutputNFTIDs{},
	)
	ViewAccountNFTsInCollection = coreutil.NewViewEP21(Contract, "accountNFTsInCollection",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamCollectionID, codec.NFTID),
		OutputNFTIDs{},
	)
	ViewNFTIDbyMintID = coreutil.NewViewEP11(Contract, "NFTIDbyMintID",
		coreutil.FieldWithCodec(ParamMintID, codec.Bytes),
		coreutil.FieldWithCodec(ParamNFTID, codec.NFTID),
	)
	ViewBalance = coreutil.NewViewEP11(Contract, "balance",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		OutputFungibleTokens{},
	)
	ViewBalanceBaseToken = coreutil.NewViewEP11(Contract, "balanceBaseToken",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamBalance, codec.Uint64),
	)
	ViewBalanceBaseTokenEVM = coreutil.NewViewEP11(Contract, "balanceBaseTokenEVM",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamBalance, codec.BigIntAbs),
	)
	ViewBalanceNativeToken = coreutil.NewViewEP21(Contract, "balanceNativeToken",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamNativeTokenID, codec.NativeTokenID),
		coreutil.FieldWithCodec(ParamBalance, codec.BigIntAbs),
	)
	ViewNativeToken = coreutil.NewViewEP11(Contract, "nativeToken",
		coreutil.FieldWithCodec(ParamFoundrySN, codec.Uint32),
		coreutil.FieldWithCodec(ParamFoundryOutputBin, codec.Output),
	)

	ViewGetAccountNonce = coreutil.NewViewEP11(Contract, "getAccountNonce",
		coreutil.FieldWithCodecOptional(ParamAgentID, codec.AgentID),
		coreutil.FieldWithCodec(ParamAccountNonce, codec.Uint64),
	)
	ViewGetNativeTokenIDRegistry = coreutil.NewViewEP01(Contract, "getNativeTokenIDRegistry",
		OutputNativeTokenIDs{},
	)
	ViewNFTData = coreutil.NewViewEP11(Contract, "nftData",
		coreutil.FieldWithCodec(ParamNFTID, codec.NFTID),
		coreutil.FieldWithCodec(ParamNFTData, codec.NewCodecEx(isc.NFTFromBytes)),
	)
	ViewTotalAssets = coreutil.NewViewEP01(Contract, "totalAssets",
		OutputFungibleTokens{},
	)
)

// request parameters
const (
	ParamAccountNonce           = "n"
	ParamAgentID                = "a"
	ParamBalance                = "B"
	ParamCollectionID           = "C"
	ParamDestroyTokens          = "y"
	ParamForceMinimumBaseTokens = "f"
	ParamFoundryOutputBin       = "b"
	ParamFoundrySN              = "s"
	ParamTokenName              = "tn"
	ParamTokenTickerSymbol      = "ts"
	ParamTokenDecimals          = "td"
	ParamGasReserve             = "g"
	ParamNFTAmount              = "A"
	ParamNFTData                = "e"
	ParamNFTID                  = "z"
	ParamNFTIDs                 = "i"
	ParamNFTImmutableData       = "I"
	ParamNFTWithdrawOnMint      = "w"
	ParamMintID                 = "D"
	ParamNativeTokenID          = "N"
	ParamSupplyDeltaAbs         = "d"
	ParamTokenScheme            = "t"
)

type EPNativeTokenCreate struct {
	coreutil.EntryPointInfo[isc.Sandbox]
}

func (e EPNativeTokenCreate) Message(
	metadata *isc.IRC30NativeTokenMetadata,
	optionalTokenScheme *iotago.TokenScheme,
) isc.Message {
	params := dict.Dict{
		ParamTokenName:         codec.String.Encode(metadata.Name),
		ParamTokenTickerSymbol: codec.String.Encode(metadata.Symbol),
		ParamTokenDecimals:     codec.Uint8.Encode(metadata.Decimals),
	}
	if optionalTokenScheme != nil {
		params[ParamTokenScheme] = codec.TokenScheme.Encode(*optionalTokenScheme)
	}
	return e.EntryPointInfo.Message(params)
}

func (e EPNativeTokenCreate) WithHandler(f func(isc.Sandbox, *isc.IRC30NativeTokenMetadata, *iotago.TokenScheme) uint32) *coreutil.EntryPointHandler[isc.Sandbox] {
	return e.EntryPointInfo.WithHandler(func(ctx isc.Sandbox) dict.Dict {
		params := ctx.Params().Dict
		tokenName := codec.String.MustDecode(params[ParamTokenName])
		tokenTickerSymbol := codec.String.MustDecode(params[ParamTokenTickerSymbol])
		tokenDecimals := codec.Uint8.MustDecode(params[ParamTokenDecimals])
		metadata := isc.NewIRC30NativeTokenMetadata(tokenName, tokenTickerSymbol, tokenDecimals)
		var tokenScheme *iotago.TokenScheme
		if params[ParamTokenScheme] != nil {
			ts := codec.TokenScheme.MustDecode(params[ParamTokenScheme])
			tokenScheme = &ts
		}

		sn := f(ctx, metadata, tokenScheme)
		return dict.Dict{ParamFoundrySN: codec.Uint32.Encode(sn)}
	})
}

type EPNativeTokenModifySupply struct {
	coreutil.EntryPointInfo[isc.Sandbox]
}

func (e EPNativeTokenModifySupply) MintTokens(foundrySN uint32, delta *big.Int) isc.Message {
	return e.EntryPointInfo.Message(dict.Dict{
		ParamFoundrySN:      codec.Uint32.Encode(foundrySN),
		ParamSupplyDeltaAbs: codec.BigIntAbs.Encode(delta),
	})
}

func (e EPNativeTokenModifySupply) DestroyTokens(foundrySN uint32, delta *big.Int) isc.Message {
	return e.MintTokens(foundrySN, delta).
		WithParam(ParamDestroyTokens, codec.Bool.Encode(true))
}

func (e EPNativeTokenModifySupply) WithHandler(f func(isc.Sandbox, uint32, *big.Int, bool)) *coreutil.EntryPointHandler[isc.Sandbox] {
	return e.EntryPointInfo.WithHandler(func(ctx isc.Sandbox) dict.Dict {
		d := ctx.Params().Dict
		sn := lo.Must(codec.Uint32.Decode(d[ParamFoundrySN]))
		delta := lo.Must(codec.BigIntAbs.Decode(d[ParamSupplyDeltaAbs]))
		destroy := lo.Must(codec.Bool.Decode(d[ParamDestroyTokens], false))
		f(ctx, sn, delta, destroy)
		return nil
	})
}

type EPMintNFT struct {
	coreutil.EntryPointInfo[isc.Sandbox]
}

type EPMintNFTMessage struct{ isc.Message }

func (e EPMintNFT) Message(
	immutableMetadata []byte,
	target isc.AgentID,
	withdrawOnMint *bool,
	collectionID *iotago.NFTID,
) isc.Message {
	params := dict.Dict{
		ParamNFTImmutableData: immutableMetadata,
		ParamAgentID:          codec.AgentID.Encode(target),
	}
	if withdrawOnMint != nil {
		params[ParamNFTWithdrawOnMint] = codec.Bool.Encode(*withdrawOnMint)
	}
	if collectionID != nil {
		params[ParamCollectionID] = codec.NFTID.Encode(*collectionID)
	}
	return e.EntryPointInfo.Message(params)
}

func (e EPMintNFT) WithHandler(f func(isc.Sandbox, []byte, isc.AgentID, bool, iotago.NFTID) []byte) *coreutil.EntryPointHandler[isc.Sandbox] {
	return e.EntryPointInfo.WithHandler(func(ctx isc.Sandbox) dict.Dict {
		d := ctx.Params().Dict
		immutableMetadata := lo.Must(codec.Bytes.Decode(d[ParamNFTImmutableData]))
		target := lo.Must(codec.AgentID.Decode(d[ParamAgentID]))
		withdraw := lo.Must(codec.Bool.Decode(d[ParamNFTWithdrawOnMint], false))
		collID := lo.Must(codec.NFTID.Decode(d[ParamCollectionID], iotago.NFTID{}))

		mintID := f(ctx, immutableMetadata, target, withdraw, collID)
		return dict.Dict{ParamMintID: mintID}
	})
}

type OutputNFTIDs struct{}

func (OutputNFTIDs) Encode(nftIDs []iotago.NFTID) dict.Dict {
	// TODO: add pagination?
	if len(nftIDs) > math.MaxUint16 {
		panic("too many NFTs")
	}
	return codec.SliceToArray(codec.NFTID, nftIDs, ParamNFTIDs)
}

func (OutputNFTIDs) Decode(r dict.Dict) ([]iotago.NFTID, error) {
	return codec.SliceFromArray(codec.NFTID, r, ParamNFTIDs)
}

type OutputSerialNumberSet struct{}

func (OutputSerialNumberSet) Encode(sns map[uint32]struct{}) dict.Dict {
	return codec.SliceToDictKeys(codec.Uint32, lo.Keys(sns))
}

func (OutputSerialNumberSet) Has(r dict.Dict, sn uint32) bool {
	return r.Has(kv.Key(codec.Uint32.Encode(sn)))
}

func (OutputSerialNumberSet) Decode(r dict.Dict) (map[uint32]struct{}, error) {
	sns, err := codec.SliceFromDictKeys(codec.Uint32, r)
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(sns, func(sn uint32) (uint32, struct{}) { return sn, struct{}{} }), nil
}

type OutputNativeTokenIDs struct{}

func (OutputNativeTokenIDs) Encode(ids []isc.NativeTokenID) dict.Dict {
	return codec.SliceToDictKeys(codec.NativeTokenID, ids)
}

func (OutputNativeTokenIDs) Decode(r dict.Dict) ([]isc.NativeTokenID, error) {
	return codec.SliceFromDictKeys(codec.NativeTokenID, r)
}

type OutputFungibleTokens struct{}

func (OutputFungibleTokens) Encode(fts *isc.Assets) dict.Dict {
	return fts.ToDict()
}

func (OutputFungibleTokens) Decode(r dict.Dict) (*isc.Assets, error) {
	return isc.AssetsFromDict(r)
}

type OutputAccountList struct{ coreutil.RawDictCodec }

func (OutputAccountList) DecodeAccounts(allAccounts dict.Dict, chainID isc.ChainID) ([]isc.AgentID, error) {
	return codec.SliceFromDictKeys(
		codec.NewCodecEx(func(b []byte) (isc.AgentID, error) { return agentIDFromKey(kv.Key(b), chainID) }),
		allAccounts,
	)
}
