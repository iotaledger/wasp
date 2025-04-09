// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// ChainInfo is an API structure containing the main parameters of the chain
type ChainInfo struct {
	ChainID         ChainID
	ChainAdmin      AgentID
	GasFeePolicy    *gas.FeePolicy
	GasLimits       *gas.Limits
	BlockKeepAmount int32

	PublicURL string
	Metadata  *PublicChainMetadata
}

func (c *ChainInfo) Bytes() []byte {
	return bcs.MustMarshal(c)
}

func ChainInfoFromBytes(b []byte) (*ChainInfo, error) {
	return bcs.Unmarshal[*ChainInfo](b)
}
