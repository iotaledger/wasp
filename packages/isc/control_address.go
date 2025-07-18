// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type ControlAddresses struct {
	AnchorOwner     *cryptolib.Address
	ChainAdmin      AgentID
	SinceBlockIndex uint32
}
