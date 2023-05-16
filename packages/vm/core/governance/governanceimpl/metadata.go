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

// MaxMetadataLength - Version - SchemaVersion - L1Commitment - GasFeePolicy - CustomMetadataLength
const MaxCustomMetadataLength = iotago.MaxMetadataLength - serializer.OneByte - serializer.UInt32ByteSize - state.L1CommitmentSize - gas.GasPolicyByteSize - serializer.UInt16ByteSize

func setMetadata(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	publicURLBytes := ctx.Params().Get(governance.ParamPublicURL)
	publicURL, err := codec.DecodeString(publicURLBytes)
	ctx.RequireNoError(err)

	evmJSONRPCUrlBytes := ctx.Params().Get(governance.ParamMetadataEVMJsonRPCURL)
	evmJSONRPCUrl, err := codec.DecodeString(evmJSONRPCUrlBytes, "")
	ctx.RequireNoError(err)

	evmWebSocketURLBytes := ctx.Params().Get(governance.ParamMetadataEVMWebSocketURL)
	evmWebSocketURL, err := codec.DecodeString(evmWebSocketURLBytes, "")
	ctx.RequireNoError(err)

	ctx.Requiref(len(publicURL) <= MaxCustomMetadataLength, "custom metadata size too big (%d>%d)", len(publicURL), MaxCustomMetadataLength)

	governance.SetPublicURL(ctx.State(), publicURL)
	governance.SetEVMJsonRPCURL(ctx.State(), evmJSONRPCUrl)
	governance.SetEVMWebSocketURL(ctx.State(), evmWebSocketURL)

	return nil
}

func getMetadata(ctx isc.SandboxView) dict.Dict {
	publicURL, err := governance.GetPublicURL(ctx.StateR())
	ctx.RequireNoError(err)

	evmJSONRPCUrl, err := governance.GetEVMJsonRPCURL(ctx.StateR())
	ctx.RequireNoError(err)

	evmWebSocketURL, err := governance.GetEVMWebSocketURL(ctx.StateR())
	ctx.RequireNoError(err)

	return dict.Dict{
		governance.ParamPublicURL:               []byte(publicURL),
		governance.ParamMetadataEVMJsonRPCURL:   []byte(evmJSONRPCUrl),
		governance.ParamMetadataEVMWebSocketURL: []byte(evmWebSocketURL),
	}
}
