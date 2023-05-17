package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

const MaxURLLength = 255

func setMetadata(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	publicURLBytes := ctx.Params().Get(governance.ParamPublicURL)
	publicURL, err := codec.DecodeString(publicURLBytes, "")
	ctx.RequireNoError(err)
	ctx.Requiref(len(publicURL) <= MaxURLLength, "public url size too big (%d>%d)", len(publicURL), MaxURLLength)

	evmJSONRPCUrlBytes := ctx.Params().Get(governance.ParamMetadataEVMJsonRPCURL)
	evmJSONRPCUrl, err := codec.DecodeString(evmJSONRPCUrlBytes, "")
	ctx.RequireNoError(err)
	ctx.Requiref(len(evmJSONRPCUrl) <= MaxURLLength, "evm json rpc url size too big (%d>%d)", len(evmJSONRPCUrl), MaxURLLength)

	evmWebSocketURLBytes := ctx.Params().Get(governance.ParamMetadataEVMWebSocketURL)
	evmWebSocketURL, err := codec.DecodeString(evmWebSocketURLBytes, "")
	ctx.RequireNoError(err)
	ctx.Requiref(len(evmWebSocketURL) <= MaxURLLength, "evm websocket url size too big (%d>%d)", len(evmWebSocketURL), MaxURLLength)

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
