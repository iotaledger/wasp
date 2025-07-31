// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// handler for ISCUtil::hn
func (h *magicContractHandler) Hn(s string) isc.Hname {
	return isc.Hn(s)
}

// handler for ISCUtil::print
func (h *magicContractHandler) Print(s string) {
	h.ctx.Log().Debugf("ISCUtil::print -> %q", s)
}
