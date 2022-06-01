package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
)

// BasicOutputFromPostData creates extended output object from parameters.
// It automatically adjusts amount of iotas required for the dust deposit
func BasicOutputFromPostData(
	senderAddress iotago.Address,
	senderContract iscp.Hname,
	par iscp.RequestParameters,
) *iotago.BasicOutput {
	metadata := par.Metadata
	if metadata == nil {
		// if metadata is not specified, target is nil. It corresponds to sending funds to the plain L1 address
		metadata = &iscp.SendMetadata{}
	}

	ret := MakeBasicOutput(
		par.TargetAddress,
		senderAddress,
		par.FungibleTokens,
		&iscp.RequestMetadata{
			SenderContract: senderContract,
			TargetContract: metadata.TargetContract,
			EntryPoint:     metadata.EntryPoint,
			Params:         metadata.Params,
			Allowance:      metadata.Allowance,
			GasBudget:      metadata.GasBudget,
		},
		par.Options,
		!par.AdjustToMinimumDustDeposit,
	)
	return ret
}

// MakeBasicOutput creates new output from input parameters.
// Auto adjusts minimal dust deposit if the notAutoAdjust flag is absent or false
// If auto adjustment to dust is disabled and not enough iotas, returns an error
func MakeBasicOutput(
	targetAddress iotago.Address,
	senderAddress iotago.Address,
	assets *iscp.FungibleTokens,
	metadata *iscp.RequestMetadata,
	options iscp.SendOptions,
	disableAutoAdjustDustDeposit ...bool,
) *iotago.BasicOutput {
	if assets == nil {
		assets = &iscp.FungibleTokens{}
	}
	out := &iotago.BasicOutput{
		Amount:       assets.Iotas,
		NativeTokens: assets.Tokens,
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
	if options.Timelock != nil {
		cond := &iotago.TimelockUnlockCondition{
			MilestoneIndex: options.Timelock.MilestoneIndex,
		}
		if !options.Timelock.Time.IsZero() {
			cond.UnixTime = uint32(options.Timelock.Time.Unix())
		}
		out.Conditions = append(out.Conditions, cond)
	}
	if options.Expiration != nil {
		cond := &iotago.ExpirationUnlockCondition{
			MilestoneIndex: options.Expiration.MilestoneIndex,
			ReturnAddress:  options.Expiration.ReturnAddress,
		}
		if !options.Expiration.Time.IsZero() {
			cond.UnixTime = uint32(options.Expiration.Time.Unix())
		}
		out.Conditions = append(out.Conditions, cond)
	}

	// Adjust to minimum dust deposit, if needed
	if len(disableAutoAdjustDustDeposit) > 0 && disableAutoAdjustDustDeposit[0] {
		return out
	}

	storageDeposit := parameters.L1.Protocol.RentStructure.MinRent(out)
	if out.Deposit() < storageDeposit {
		// adjust the amount to the minimum required
		out.Amount = storageDeposit
	}

	return out
}

func NFTOutputFromPostData(
	senderAddress iotago.Address,
	senderContract iscp.Hname,
	par iscp.RequestParameters,
	nft *iscp.NFT,
) *iotago.NFTOutput {
	basicOutput := BasicOutputFromPostData(senderAddress, senderContract, par)
	out := NftOutputFromBasicOutput(basicOutput, nft)

	if !par.AdjustToMinimumDustDeposit {
		return out
	}
	storageDeposit := parameters.L1.Protocol.RentStructure.MinRent(out)
	if out.Deposit() < storageDeposit {
		// adjust the amount to the minimum required
		out.Amount = storageDeposit
	}
	return out
}

func NftOutputFromBasicOutput(o *iotago.BasicOutput, nft *iscp.NFT) *iotago.NFTOutput {
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

func AssetsFromOutput(o iotago.Output) *iscp.FungibleTokens {
	return &iscp.FungibleTokens{
		Iotas:  o.Deposit(),
		Tokens: o.NativeTokenSet(),
	}
}
