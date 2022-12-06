// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputMempoolProposal struct {
	requestRefs []*isc.RequestRef
}

func NewInputMempoolProposal(requestRefs []*isc.RequestRef) gpa.Input {
	return &inputMempoolProposal{requestRefs: requestRefs}
}
