// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

// ChainInfo is an API structure which contains main properties of the chain in on place
type ChainInfo struct {
	ChainID             *iscp.ChainID
	ChainOwnerID        *iscp.AgentID
	Description         string
	FeeAssetID          []byte
	DefaultOwnerFee     int64
	DefaultValidatorFee int64
	MaxBlobSize         uint32
	MaxEventSize        uint16
	MaxEventsPerReq     uint16
}
