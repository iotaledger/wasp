// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/coin"
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
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), sender),
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), receiver),
		allowance.Unwrap(),
	)
}

// handler for ISCPrivileged::setAllowanceBaseTokens
func (h *magicContractHandler) SetAllowanceBaseTokens(
	from common.Address,
	to common.Address,
	amount coin.Value,
) {
	setAllowanceBaseTokens(h.ctx, from, to, amount)
}

// handler for ISCPrivileged::setAllowanceCoin
func (h *magicContractHandler) SetAllowanceCoin(
	from common.Address,
	to common.Address,
	coinType coin.Type,
	amount coin.Value,
) {
	setAllowanceCoin(h.ctx, from, to, coinType, amount)
}

// handler for ISCPrivileged::moveAllowedFunds
func (h *magicContractHandler) MoveAllowedFunds(
	from common.Address,
	to common.Address,
	allowance iscmagic.ISCAssets,
) {
	assets := allowance.Unwrap()
	subtractFromAllowance(h.ctx, from, to, assets)
	h.ctx.Privileged().MustMoveBetweenAccounts(
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), from),
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), to),
		assets,
	)
}
