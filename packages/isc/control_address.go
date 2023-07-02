// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import iotago "github.com/iotaledger/iota.go/v3"

type ControlAddresses struct {
	StateAddress     iotago.Address
	GoverningAddress iotago.Address
	SinceBlockIndex  uint32
}
