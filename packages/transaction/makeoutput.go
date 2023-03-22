package transaction

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
)

// BasicOutputFromPostData creates extended output object from parameters.
// It automatically adjusts amount of base tokens required for the storage deposit
func BasicOutputFromPostData(
	senderAddress iotago.Address,
	senderContract isc.Hname,
	par isc.RequestParameters,
) *iotago.BasicOutput {
	metadata := par.Metadata
	if metadata == nil {
		// if metadata is not specified, target is nil. It corresponds to sending funds to the plain L1 address
		metadata = &isc.SendMetadata{}
	}

	ret := MakeBasicOutput(
		par.TargetAddress,
		senderAddress,
		par.Assets,
		&isc.RequestMetadata{
			SenderContract: senderContract,
			TargetContract: metadata.TargetContract,
			EntryPoint:     metadata.EntryPoint,
			Params:         metadata.Params,
			Allowance:      metadata.Allowance,
			GasBudget:      metadata.GasBudget,
		},
		par.Options,
	)
	if par.AdjustToMinimumStorageDeposit {
		return AdjustToMinimumStorageDeposit(ret)
	}
	return ret
}

// MakeBasicOutput creates new output from input parameters (ignoring storage deposit).
func MakeBasicOutput(
	targetAddress iotago.Address,
	senderAddress iotago.Address,
	assets *isc.Assets,
	metadata *isc.RequestMetadata,
	options isc.SendOptions,
) *iotago.BasicOutput {
	if assets == nil {
		assets = &isc.Assets{}
	}
	out := &iotago.BasicOutput{
		Amount:       assets.BaseTokens,
		NativeTokens: assets.NativeTokens,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddress},
		},
	}
	if senderAddress != nil {
		out.Features = append(out.Features, &iotago.SenderFeature{
			Address: senderAddress,
		})
	}
	if metadata != nil {
		out.Features = append(out.Features, &iotago.MetadataFeature{
			Data: metadata.Bytes(),
		})
	}
	if !options.Timelock.IsZero() {
		cond := &iotago.TimelockUnlockCondition{
			UnixTime: uint32(options.Timelock.Unix()),
		}
		out.Conditions = append(out.Conditions, cond)
	}
	if options.Expiration != nil {
		cond := &iotago.ExpirationUnlockCondition{
			ReturnAddress: options.Expiration.ReturnAddress,
		}
		if !options.Expiration.Time.IsZero() {
			cond.UnixTime = uint32(options.Expiration.Time.Unix())
		}
		out.Conditions = append(out.Conditions, cond)
	}
	return out
}

func NFTOutputFromPostData(
	senderAddress iotago.Address,
	senderContract isc.Hname,
	par isc.RequestParameters,
	nft *isc.NFT,
) *iotago.NFTOutput {
	basicOutput := BasicOutputFromPostData(senderAddress, senderContract, par)
	out := NftOutputFromBasicOutput(basicOutput, nft)

	if !par.AdjustToMinimumStorageDeposit {
		return out
	}
	storageDeposit := parameters.L1().Protocol.RentStructure.MinRent(out)
	if out.Deposit() < storageDeposit {
		// adjust the amount to the minimum required
		out.Amount = storageDeposit
	}
	return out
}

func NftOutputFromBasicOutput(o *iotago.BasicOutput, nft *isc.NFT) *iotago.NFTOutput {
	return &iotago.NFTOutput{
		Amount:       o.Amount,
		NativeTokens: o.NativeTokens,
		Features:     o.Features,
		Conditions:   o.Conditions,
		NFTID:        nft.ID,
		ImmutableFeatures: iotago.Features{
			&iotago.IssuerFeature{Address: nft.Issuer},
			&iotago.MetadataFeature{Data: nft.Metadata},
		},
	}
}

func AssetsFromOutput(o iotago.Output) *isc.Assets {
	return &isc.Assets{
		BaseTokens:   o.Deposit(),
		NativeTokens: o.NativeTokenList(),
	}
}

func AdjustToMinimumStorageDeposit[T iotago.Output](out T) T {
	storageDeposit := parameters.L1().Protocol.RentStructure.MinRent(out)
	if out.Deposit() >= storageDeposit {
		return out
	}
	switch out := iotago.Output(out).(type) {
	case *iotago.AliasOutput:
		out.Amount = storageDeposit
	case *iotago.BasicOutput:
		out.Amount = storageDeposit
	case *iotago.FoundryOutput:
		out.Amount = storageDeposit
	case *iotago.NFTOutput:
		out.Amount = storageDeposit
	default:
		panic(fmt.Sprintf("no handler for output type %T", out))
	}
	return out
}
