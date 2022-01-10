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
	ret := MakeExtendedOutput(
		par.TargetAddress,
		senderAddress,
		par.Assets,
		&iscp.RequestMetadata{
			SenderContract: senderContract,
			TargetContract: par.Metadata.TargetContract,
			EntryPoint:     par.Metadata.EntryPoint,
			Params:         par.Metadata.Params,
			Allowance:      par.Metadata.Allowance,
			GasBudget:      par.Metadata.GasBudget,
		},
		par.Options,
		rentStructure,
	)
	return ret
}

// MakeExtendedOutput creates new ExtendedOutput from input parameters.
// Adjusts dust deposit if needed and returns flag if adjusted
func MakeExtendedOutput(
	targetAddress iotago.Address,
	senderAddress iotago.Address,
	assets *iscp.Assets,
	metadata *iscp.RequestMetadata,
	options *iscp.SendOptions,
	rentStructure *iotago.RentStructure,
) *iotago.ExtendedOutput {
	if assets == nil {
		assets = &iscp.Assets{}
	}
	ret := &iotago.ExtendedOutput{
		Address:      targetAddress,
		Amount:       assets.Iotas,
		NativeTokens: assets.Tokens,
		Blocks:       iotago.FeatureBlocks{},
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

	if options != nil {
		panic(" send options FeatureBlocks not implemented yet")
	}

	// Adjust to minimum dust deposit
	neededDustDeposit := ret.VByteCost(rentStructure, nil)
	if ret.Amount < neededDustDeposit {
		ret.Amount = neededDustDeposit
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
