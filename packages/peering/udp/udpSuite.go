// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
)

type Suite interface {
	pairing.Suite
	kyber.Group
}
