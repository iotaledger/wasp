package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

// ExtendedOutputFromPostData creates extended output object from parameters.
// It automatically adjusts amount of iotas required for the dust deposit
func ExtendedOutputFromPostData(
	senderAddress iotago.Address,
	senderContract iscp.Hname,
	par iscp.RequestParameters,
	rentStructure *iotago.RentStructure,
) *iotago.ExtendedOutput {
	metadata := par.Metadata
	if metadata == nil {
		// if metadata is not specified, target is nil. It corresponds to sending funds to the plain L1 address
		metadata = &iscp.SendMetadata{}
	}
	ret := MakeExtendedOutput(
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
		par.Options,
		rentStructure,
		!par.AdjustToMinimumDustDeposit,
	)
	return ret
}

// MakeExtendedOutput creates new ExtendedOutput from input parameters.
// Auto adjusts minimal dust deposit if the notAutoAdjust flag is absent or false
// If auto adjustment to dust is disabled and not enough iotas, returns an error
func MakeExtendedOutput(
	targetAddress iotago.Address,
	senderAddress iotago.Address,
	assets *iscp.Assets,
	metadata *iscp.RequestMetadata,
	options iscp.SendOptions,
	rentStructure *iotago.RentStructure,
	disableAutoAdjustDustDeposit ...bool,
) *iotago.ExtendedOutput {
	if assets == nil {
		assets = &iscp.Assets{}
	}
	ret := &iotago.ExtendedOutput{
		Amount:       assets.Iotas,
		NativeTokens: assets.Tokens,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddress},
		},
	}
	if senderAddress != nil {
		ret.Blocks = append(ret.Blocks, &iotago.SenderFeatureBlock{
			Address: senderAddress,
		})
	}
	if metadata != nil {
		ret.Blocks = append(ret.Blocks, &iotago.MetadataFeatureBlock{
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
		ret.Conditions = append(ret.Conditions, cond)
	}
	if options.Expiration != nil {
		cond := &iotago.ExpirationUnlockCondition{
			MilestoneIndex: options.Expiration.MilestoneIndex,
			ReturnAddress:  options.Expiration.ReturnAddress,
		}
		if !options.Expiration.Time.IsZero() {
			cond.UnixTime = uint32(options.Expiration.Time.Unix())
		}
		ret.Conditions = append(ret.Conditions, cond)
	}

	// Adjust to minimum dust deposit, if needed
	if len(disableAutoAdjustDustDeposit) > 0 && disableAutoAdjustDustDeposit[0] {
		return ret
	}
	requiredDustDeposit := ret.VByteCost(rentStructure, nil)
	if ret.Amount < requiredDustDeposit {
		// adjust the amount to the minimum required
		ret.Amount = requiredDustDeposit
	}
	return ret
}

func AssetsFromOutput(o iotago.Output) *iscp.Assets {
	switch o := o.(type) {
	case *iotago.ExtendedOutput:
		return &iscp.Assets{
			Iotas:  o.Amount,
			Tokens: o.NativeTokens,
		}
	default:
		panic(xerrors.Errorf("AssetsFromExtendedOutput: not supported output type: %T", o))
	}
}
