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

const (
	MaxCustomMetadataLength = iotago.MaxMetadataLength - serializer.OneByte - serializer.UInt32ByteSize - state.L1CommitmentSize - gas.FeePolicyByteSize - serializer.UInt16ByteSize
)

func setMetadata(ctx isc.Sandbox, publicURLOpt *string, metadataOpt **isc.PublicChainMetadata) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	state := governance.NewStateWriterFromSandbox(ctx)
	n := 0
	if publicURLOpt != nil {
		n += len(*publicURLOpt)
		state.SetPublicURL(*publicURLOpt)
	}
	if metadataOpt != nil {
		n += len((*metadataOpt).Bytes())
		state.SetMetadata(*metadataOpt)
	}
	ctx.Requiref(n <= MaxCustomMetadataLength, "supplied publicURL and metadata is too big (%d>%d)", n, MaxCustomMetadataLength)

	return nil
}

func getMetadata(ctx isc.SandboxView) (string, *isc.PublicChainMetadata) {
	state := governance.NewStateReaderFromSandbox(ctx)
	publicURL := state.GetPublicURL()
	metadata := state.GetMetadata()
	return publicURL, metadata
}
