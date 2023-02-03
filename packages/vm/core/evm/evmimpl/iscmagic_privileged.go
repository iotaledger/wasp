// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCPrivileged::moveBetweenAccounts
func (h *magicContractHandler) MoveBetweenAccounts(
	sender common.Address,
	receiver common.Address,
	allowance iscmagic.ISCAssets,
) {
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(sender),
		isc.NewEthereumAddressAgentID(receiver),
		allowance.Unwrap(),
	)
}

// handler for ISCPrivileged::addToAllowance
func (h *magicContractHandler) AddToAllowance(
	from common.Address,
	to common.Address,
	allowance iscmagic.ISCAssets,
) {
	addToAllowance(h.ctx, from, to, allowance.Unwrap())
}

// handler for ISCPrivileged::moveAllowedFunds
func (h *magicContractHandler) MoveAllowedFunds(
	from common.Address,
	to common.Address,
	allowance iscmagic.ISCAssets,
) {
	taken := subtractFromAllowance(h.ctx, from, to, allowance.Unwrap())
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(from),
		isc.NewEthereumAddressAgentID(to),
		taken,
	)
}
