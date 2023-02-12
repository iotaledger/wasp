// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputMempoolProposal struct {
	requestRefs []*isc.RequestRef
}

func NewInputMempoolProposal(requestRefs []*isc.RequestRef) gpa.Input {
	return &inputMempoolProposal{requestRefs: requestRefs}
}

func (inp *inputMempoolProposal) String() string {
	acc := "{cons.inputMempoolProposal: "
	for i, ref := range inp.requestRefs {
		if i > 3 {
			acc += fmt.Sprintf("..., %v in total", len(inp.requestRefs))
			break
		}
		acc += fmt.Sprintf(" %v", ref.String())
	}
	acc += "}"
	return acc
}
