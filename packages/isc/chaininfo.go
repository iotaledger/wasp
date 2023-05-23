// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type ChainMetadata struct {
	EVMJsonRPCURL   string
	EVMWebSocketURL string

	ChainName        string
	ChainDescription string
	ChainOwnerEmail  string
	ChainWebsite     string
}

// ChainInfo is an API structure containing the main parameters of the chain
type ChainInfo struct {
	ChainID         ChainID
	ChainOwnerID    AgentID
	GasFeePolicy    *gas.FeePolicy
	GasLimits       *gas.Limits
	BlockKeepAmount int32

	PublicURL string
	Metadata  ChainMetadata
}
