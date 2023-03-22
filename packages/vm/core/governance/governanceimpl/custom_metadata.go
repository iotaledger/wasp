package governanceimpl

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// MaxMetadataLength - Version - SchemaVersion - L1Commitment - GasFeePolicy - CustomMetadataLength
const MaxCustomMetadataLength = iotago.MaxMetadataLength - serializer.OneByte - serializer.UInt32ByteSize - state.L1CommitmentSize - gas.GasPolicyByteSize - serializer.UInt16ByteSize

func setCustomMetadata(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	customMetadata := ctx.Params().MustGet(governance.ParamCustomMetadata)
	ctx.Requiref(len(customMetadata) <= MaxCustomMetadataLength, "custom metadata size too big (%d>%d)", len(customMetadata), MaxCustomMetadataLength)
	governance.SetCustomMetadata(ctx.State(), customMetadata)
	return nil
}

func getCustomMetadata(ctx isc.SandboxView) dict.Dict {
	return dict.Dict{
		governance.ParamCustomMetadata: governance.GetCustomMetadata(ctx.StateR()),
	}
}
