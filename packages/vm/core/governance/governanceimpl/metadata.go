package governanceimpl

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	MaxCustomMetadataLength = iotago.MaxMetadataLength - serializer.OneByte - serializer.UInt32ByteSize - state.L1CommitmentSize - gas.FeePolicyByteSize - serializer.UInt16ByteSize
)

func setMetadata(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	var publicURLBytes []byte
	var metadataBytes []byte

	publicURLBytes = ctx.Params().Get(governance.ParamPublicURL)
	metadataBytes = ctx.Params().Get(governance.ParamMetadata)

	ctx.Requiref(len(publicURLBytes)+len(metadataBytes) <= MaxCustomMetadataLength, "supplied publicUrl and metadata is too big (%d>%d)", len(publicURLBytes)+len(metadataBytes), MaxCustomMetadataLength)

	if publicURLBytes != nil {
		publicURL, err := codec.DecodeString(publicURLBytes, "")
		ctx.RequireNoError(err)
		governance.SetPublicURL(ctx.State(), publicURL)
	}

	if metadataBytes != nil {
		metadata, err := isc.PublicChainMetadataFromBytes(metadataBytes)
		ctx.RequireNoError(err)
		governance.SetMetadata(ctx.State(), metadata)
	}

	return nil
}

func getMetadata(ctx isc.SandboxView) dict.Dict {
	publicURL, _ := governance.GetPublicURL(ctx.StateR())
	metadata := governance.MustGetMetadata(ctx.StateR())

	return dict.Dict{
		governance.ParamPublicURL: []byte(publicURL),
		governance.ParamMetadata:  metadata.Bytes(),
	}
}
