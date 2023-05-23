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
	MaxCustomMetadataLength   = iotago.MaxMetadataLength - serializer.OneByte - serializer.UInt32ByteSize - state.L1CommitmentSize - gas.GasPolicyByteSize - serializer.UInt16ByteSize
	MaxURLLength              = 1000
	MaxChainNameLength        = 255
	MaxChainDescriptionLength = 2048
	MaxChainOwnerEmailLength  = 255
)

func setMetadata(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	publicURLBytes := ctx.Params().Get(governance.ParamPublicURL)
	if publicURLBytes != nil {
		publicURL, err := codec.DecodeString(publicURLBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(publicURL) <= MaxCustomMetadataLength, "public url size too big (%d>%d)", len(publicURL), MaxCustomMetadataLength)
		governance.SetPublicURL(ctx.State(), publicURL)
	}

	evmJSONRPCUrlBytes := ctx.Params().Get(governance.ParamMetadataEVMJsonRPCURL)
	if evmJSONRPCUrlBytes != nil {
		evmJSONRPCUrl, err := codec.DecodeString(evmJSONRPCUrlBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(evmJSONRPCUrl) <= MaxURLLength, "evm json rpc url size too big (%d>%d)", len(evmJSONRPCUrl), MaxURLLength)
		governance.SetEVMJsonRPCURL(ctx.State(), evmJSONRPCUrl)
	}

	evmWebSocketURLBytes := ctx.Params().Get(governance.ParamMetadataEVMWebSocketURL)
	if evmWebSocketURLBytes != nil {
		evmWebSocketURL, err := codec.DecodeString(evmWebSocketURLBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(evmWebSocketURL) <= MaxURLLength, "evm websocket url size too big (%d>%d)", len(evmWebSocketURL), MaxURLLength)
		governance.SetEVMWebSocketURL(ctx.State(), evmWebSocketURL)
	}

	chainNameBytes := ctx.Params().Get(governance.ParamMetadataChainName)
	if chainNameBytes != nil {
		chainName, err := codec.DecodeString(chainNameBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(chainName) <= MaxChainNameLength, "evm websocket url size too big (%d>%d)", len(chainName), MaxChainNameLength)
		governance.SetChainName(ctx.State(), chainName)
	}

	chainDescriptionBytes := ctx.Params().Get(governance.ParamMetadataChainDescription)
	if chainDescriptionBytes != nil {
		chainDescription, err := codec.DecodeString(chainDescriptionBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(chainDescription) <= MaxChainDescriptionLength, "evm websocket url size too big (%d>%d)", len(chainDescription), MaxChainDescriptionLength)
		governance.SetChainDescription(ctx.State(), chainDescription)
	}

	chainOwnerEmailBytes := ctx.Params().Get(governance.ParamMetadataChainOwnerEmail)
	if chainOwnerEmailBytes != nil {
		chainOwnerEmail, err := codec.DecodeString(chainOwnerEmailBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(chainOwnerEmail) <= MaxChainOwnerEmailLength, "evm websocket url size too big (%d>%d)", len(chainOwnerEmail), MaxChainOwnerEmailLength)
		governance.SetChainOwnerEmail(ctx.State(), chainOwnerEmail)
	}

	chainWebsiteBytes := ctx.Params().Get(governance.ParamMetadataChainWebsite)
	if chainWebsiteBytes != nil {
		chainWebsite, err := codec.DecodeString(chainWebsiteBytes, "")
		ctx.RequireNoError(err)
		ctx.Requiref(len(chainWebsite) <= MaxURLLength, "evm websocket url size too big (%d>%d)", len(chainWebsite), MaxURLLength)
		governance.SetChainWebsite(ctx.State(), chainWebsite)
	}

	return nil
}

func getMetadata(ctx isc.SandboxView) dict.Dict {
	publicURL, err := governance.GetPublicURL(ctx.StateR())
	ctx.RequireNoError(err)

	evmJSONRPCUrl, err := governance.GetEVMJsonRPCURL(ctx.StateR())
	ctx.RequireNoError(err)

	evmWebSocketURL, err := governance.GetEVMWebSocketURL(ctx.StateR())
	ctx.RequireNoError(err)

	chainName, err := governance.GetChainName(ctx.StateR())
	ctx.RequireNoError(err)

	chainDescription, err := governance.GetChainDescription(ctx.StateR())
	ctx.RequireNoError(err)

	chainOwnerEmail, err := governance.GetChainOwnerEmail(ctx.StateR())
	ctx.RequireNoError(err)

	chainWebsite, err := governance.GetChainWebsite(ctx.StateR())
	ctx.RequireNoError(err)

	return dict.Dict{
		governance.ParamPublicURL:                []byte(publicURL),
		governance.ParamMetadataEVMJsonRPCURL:    []byte(evmJSONRPCUrl),
		governance.ParamMetadataEVMWebSocketURL:  []byte(evmWebSocketURL),
		governance.ParamMetadataChainName:        []byte(chainName),
		governance.ParamMetadataChainDescription: []byte(chainDescription),
		governance.ParamMetadataChainOwnerEmail:  []byte(chainOwnerEmail),
		governance.ParamMetadataChainWebsite:     []byte(chainWebsite),
	}
}
