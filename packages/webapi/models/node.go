package models

import (
	"github.com/iotaledger/wasp/packages/parameters"
)

type NodeOwnerCertificateResponse struct {
	Certificate string `json:"certificate" swagger:"desc(Certificate stating the ownership. (Hex)),required"`
}

type ProtocolParameters struct {
	Epoch                 uint64 `json:"epoch" swagger:"desc(The protocol's current epoch),required"`
	ProtocolVersion       uint64 `json:"protocol_version" swagger:"desc(The protocol's version),required"`
	SystemStateVersion    uint64 `json:"system_state_version" swagger:"desc(The protocol's system_state_version),required"`
	IotaTotalSupply       uint64 `json:"iota_total_supply" swagger:"desc(The iota's total_supply),required"`
	ReferenceGasPrice     uint64 `json:"reference_gas_price" swagger:"desc(The current reference_gas_price),required"`
	EpochStartTimestampMs uint64 `json:"epoch_start_timestamp_ms" swagger:"desc(The current epoch's start_timestamp in ms),required"`
	EpochDurationMs       uint64 `json:"epoch_duration_ms" swagger:"desc(The current epoch's duration in ms),required"`
}

type L1Params struct {
	Protocol  ProtocolParameters   `json:"protocol" swagger:"desc(The protocol parameters),required"`
	BaseToken parameters.BaseToken `json:"baseToken" swagger:"desc(The base token parameters),required"`
}

func MapL1Params(l1 *parameters.L1Params) *L1Params {
	params := &L1Params{
		Protocol: ProtocolParameters{
			Epoch:                 l1.Protocol.Epoch.Uint64(),
			ProtocolVersion:       l1.Protocol.ProtocolVersion.Uint64(),
			SystemStateVersion:    l1.Protocol.SystemStateVersion.Uint64(),
			IotaTotalSupply:       l1.Protocol.IotaTotalSupply.Uint64(),
			ReferenceGasPrice:     l1.Protocol.ReferenceGasPrice.Uint64(),
			EpochStartTimestampMs: l1.Protocol.EpochStartTimestampMs.Uint64(),
			EpochDurationMs:       l1.Protocol.EpochDurationMs.Uint64(),
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
	Version    string    `json:"version" swagger:"desc(The version of the node),required"`
	PublicKey  string    `json:"publicKey" swagger:"desc(The public key of the node (Hex)),required"`
	PeeringURL string    `json:"peeringURL" swagger:"desc(The net id of the node),required"`
	L1Params   *L1Params `json:"l1Params" swagger:"desc(The L1 parameters),required"`
}
