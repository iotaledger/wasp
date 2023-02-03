package models

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
)

type NodeOwnerCertificateRequest struct {
	PublicKey    string `json:"publicKey" swagger:"desc(The public key of the node (Hex)),required"`
	OwnerAddress string `json:"ownerAddress" swagger:"desc(Node owner address. (Bech32)),required"`
}

type NodeOwnerCertificateResponse struct {
	Certificate string `json:"certificate" swagger:"desc(Certificate stating the ownership. (Hex)),required"`
}

// RentStructure defines the parameters of rent cost calculations on objects which take node resources.
type RentStructure struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32 `json:"vByteCost" swagger:"desc(The virtual byte cost),required,min(1)"`
	// Defines the factor to be used for data only fields.
	VBFactorData iotago.VByteCostFactor `json:"vByteFactorData" swagger:"desc(The virtual byte factor for data fields),required"`
	// defines the factor to be used for key/lookup generating fields.
	VBFactorKey iotago.VByteCostFactor `json:"vByteFactorKey" swagger:"desc(The virtual byte factor for key/lookup generating fields),required"`
}

type ProtocolParameters struct {
	// The version of the protocol running.
	Version byte `json:"version" swagger:"desc(The protocol version),required"`
	// The human friendly name of the network.
	NetworkName string `json:"networkName" swagger:"desc(The network name),required"`
	// The HRP prefix used for Bech32 addresses in the network.
	Bech32HRP iotago.NetworkPrefix `json:"bech32Hrp" swagger:"desc(The human readable network prefix),required"`
	// The minimum pow score of the network.
	MinPoWScore uint32 `json:"minPowScore" swagger:"desc(The minimal PoW score),required,min(1)"`
	// The below max depth parameter of the network.
	BelowMaxDepth uint8 `json:"belowMaxDepth" swagger:"desc(The networks max depth),required,min(1)"`
	// The rent structure used by given node/network.
	RentStructure RentStructure `json:"rentStructure" swagger:"desc(The rent structure of the protocol),required"`
	// TokenSupply defines the current token supply on the network.
	TokenSupply string `json:"tokenSupply" swagger:"desc(The token supply),required"`
}

type L1Params struct {
	MaxPayloadSize int                  `json:"maxPayloadSize" swagger:"desc(The max payload size),required"`
	Protocol       ProtocolParameters   `json:"protocol" swagger:"desc(The protocol parameters),required"`
	BaseToken      parameters.BaseToken `json:"baseToken" swagger:"desc(The base token parameters),required"`
}

func MapL1Params(l1 *parameters.L1Params) *L1Params {
	params := &L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Message fields without payload = max size of the payload
		MaxPayloadSize: l1.MaxPayloadSize,
		Protocol: ProtocolParameters{
			Version:     l1.Protocol.Version,
			NetworkName: l1.Protocol.NetworkName,
			Bech32HRP:   l1.Protocol.Bech32HRP,
			MinPoWScore: l1.Protocol.MinPoWScore,
			RentStructure: RentStructure{
				VByteCost:    l1.Protocol.RentStructure.VByteCost,
				VBFactorData: l1.Protocol.RentStructure.VBFactorData,
				VBFactorKey:  l1.Protocol.RentStructure.VBFactorKey,
			},
			TokenSupply: iotago.EncodeUint64(l1.Protocol.TokenSupply),
		},
		BaseToken: parameters.BaseToken{
			Name:            l1.BaseToken.Name,
			TickerSymbol:    l1.BaseToken.TickerSymbol,
			Unit:            l1.BaseToken.Unit,
			Subunit:         l1.BaseToken.Subunit,
			Decimals:        l1.BaseToken.Decimals,
			UseMetricPrefix: l1.BaseToken.UseMetricPrefix,
		},
	}

	return params
}

type VersionResponse struct {
	Version string `json:"version" swagger:"desc(The version of the node),required"`
}

type InfoResponse struct {
	Version   string    `json:"version" swagger:"desc(The version of the node),required"`
	PublicKey string    `json:"publicKey" swagger:"desc(The public key of the node (Hex)),required"`
	NetID     string    `json:"netID" swagger:"desc(The net id of the node),required"`
	L1Params  *L1Params `json:"l1Params" swagger:"desc(The L1 parameters),required"`
}
