// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// ChainInfo is an API structure containing the main parameters of the chain
type ChainInfo struct {
	ChainID         ChainID
	ChainOwnerID    AgentID
	GasFeePolicy    *gas.FeePolicy
	GasLimits       *gas.Limits
	BlockKeepAmount int32

	PublicURL string
	Metadata  *PublicChainMetadata
}

func (c *ChainInfo) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&c.ChainID)
	ww.Write(c.ChainOwnerID)
	ww.Write(c.GasFeePolicy)
	ww.Write(c.GasLimits)
	ww.WriteInt32(c.BlockKeepAmount)
	ww.WriteString(c.PublicURL)
	ww.Write(c.Metadata)
	return ww.Err
}

func (c *ChainInfo) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&c.ChainID)
	c.ChainOwnerID = AgentIDFromReader(rr)
	c.GasFeePolicy = rwutil.ReadStruct(rr, new(gas.FeePolicy))
	c.GasLimits = rwutil.ReadStruct(rr, new(gas.Limits))
	rr.Read(c.GasLimits)
	c.BlockKeepAmount = rr.ReadInt32()
	c.PublicURL = rr.ReadString()
	c.Metadata = rwutil.ReadStruct(rr, new(PublicChainMetadata))
	return rr.Err
}

func (c *ChainInfo) Bytes() []byte {
	return rwutil.WriteToBytes(c)
}

func ChainInfoFromBytes(b []byte) (*ChainInfo, error) {
	return rwutil.ReadFromBytes(b, new(ChainInfo))
}
