// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputMempoolRequests struct {
	requests []isc.Request
}

func NewInputMempoolRequests(requests []isc.Request) gpa.Input {
	return &inputMempoolRequests{requests: requests}
}

func (inp *inputMempoolRequests) String() string {
	acc := "{cons.inputMempoolRequests: "
	for i, req := range inp.requests {
		if i > 3 {
			acc += fmt.Sprintf("..., %v in total", len(inp.requests))
			break
		}
		acc += fmt.Sprintf(" {request, id=%v}", req.ID().String())
	}
	acc += "}"
	return acc
}
