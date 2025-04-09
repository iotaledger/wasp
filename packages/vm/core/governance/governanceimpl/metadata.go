package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

const (
	MaxCustomMetadataLength = 4096
)

func setMetadata(ctx isc.Sandbox, publicURLOpt *string, metadataOpt **isc.PublicChainMetadata) {
	ctx.RequireCallerIsChainAdmin()
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
}

func getMetadata(ctx isc.SandboxView) (string, *isc.PublicChainMetadata) {
	state := governance.NewStateReaderFromSandbox(ctx)
	publicURL := state.GetPublicURL()
	metadata := state.GetMetadata()
	return publicURL, metadata
}
