// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
)

// handler for ISCUtil::hn
func (h *magicContractViewHandler) Hn(s string) isc.Hname {
	return isc.Hn(s)
}

// handler for ISCUtil::print
func (h *magicContractViewHandler) Print(s string) {
	h.ctx.Log().Debugf("ISCUtil::print -> %q", s)
}
