package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

// OutputFromPostData creates extended output object from parameters.
// It automatically adjusts amount of iotas required for the dust deposit
func OutputFromPostData(
	senderAddress iotago.Address,
	senderContract iscp.Hname,
	par iscp.RequestParameters,
	rentStructure *iotago.RentStructure,
) iotago.Output {
	metadata := par.Metadata
	if metadata == nil {
		// if metadata is not specified, target is nil. It corresponds to sending funds to the plain L1 address
		metadata = &iscp.SendMetadata{}
	}
	ret := MakeOutput(
		par.TargetAddress,
		senderAddress,
		par.Assets,
		&iscp.RequestMetadata{
			SenderContract: senderContract,
			TargetContract: metadata.TargetContract,
			EntryPoint:     metadata.EntryPoint,
			Params:         metadata.Params,
			Allowance:      metadata.Allowance,
			GasBudget:      metadata.GasBudget,
		},
		par.NFTID,
		par.Options,
		rentStructure,
		!par.AdjustToMinimumDustDeposit,
	)
	return ret
}

// MakeOutput creates new output from input parameters.
// Auto adjusts minimal dust deposit if the notAutoAdjust flag is absent or false
// If auto adjustment to dust is disabled and not enough iotas, returns an error
func MakeOutput(
	targetAddress iotago.Address,
	senderAddress iotago.Address,
	assets *iscp.Assets,
	metadata *iscp.RequestMetadata,
	NFTID *iotago.NFTID,
	options iscp.SendOptions,
	rentStructure *iotago.RentStructure,
	disableAutoAdjustDustDeposit ...bool,
) iotago.Output {
	if assets == nil {
		assets = &iscp.Assets{}
	}
	out := &iotago.BasicOutput{
		Amount:       assets.Iotas,
		NativeTokens: assets.Tokens,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddress},
		},
	}
	if senderAddress != nil {
		out.Blocks = append(out.Blocks, &iotago.SenderFeatureBlock{
			Address: senderAddress,
		})
	}
	if metadata != nil {
		out.Blocks = append(out.Blocks, &iotago.MetadataFeatureBlock{
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

	var finalOutput iotago.Output = out
	if NFTID != nil {
		finalOutput = nftOutputFromBasicOutput(out, NFTID)
	}

	// Adjust to minimum dust deposit, if needed
	if len(disableAutoAdjustDustDeposit) > 0 && disableAutoAdjustDustDeposit[0] {
		return finalOutput
	}

	requiredDustDeposit := out.VByteCost(rentStructure, nil)
	if finalOutput.Deposit() < requiredDustDeposit {
		// adjust the amount to the minimum required
		if NFTID != nil {
			finalOutput.(*iotago.NFTOutput).Amount = requiredDustDeposit
		} else {
			finalOutput.(*iotago.BasicOutput).Amount = requiredDustDeposit
		}
	}

	return finalOutput
}

func nftOutputFromBasicOutput(o *iotago.BasicOutput, NFTID *iotago.NFTID) *iotago.NFTOutput {
	return &iotago.NFTOutput{
		Amount:       o.Amount,
		NativeTokens: o.NativeTokens,
		Blocks:       o.Blocks,
		Conditions:   o.Conditions,
	}
}

func AssetsFromOutput(o iotago.Output) *iscp.Assets {
	switch o := o.(type) {
	case *iotago.BasicOutput:
		return &iscp.Assets{
			Iotas:  o.Amount,
			Tokens: o.NativeTokens,
		}
	default:
		panic(xerrors.Errorf("AssetsFromBasicOutput: not supported output type: %T", o))
	}
}
