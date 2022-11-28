// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputMempoolRequests struct {
	requests []isc.Request
}

func NewInputMempoolRequests(requests []isc.Request) gpa.Input {
	return &inputMempoolRequests{requests: requests}
}
