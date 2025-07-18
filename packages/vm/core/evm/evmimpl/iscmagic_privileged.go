// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/iscmagic"
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

// handler for ISCPrivileged::setAllowanceBaseTokens
func (h *magicContractHandler) SetAllowanceBaseTokens(
	from common.Address,
	to common.Address,
	amount iscmagic.CoinValue,
) {
	setAllowanceBaseTokens(h.ctx, from, to, coin.Value(amount))
}

// handler for ISCPrivileged::setAllowanceCoin
func (h *magicContractHandler) SetAllowanceCoin(
	from common.Address,
	to common.Address,
	coinType iscmagic.CoinType,
	amount iscmagic.CoinValue,
) {
	setAllowanceCoin(h.ctx, from, to, coin.MustTypeFromString(coinType), coin.Value(amount))
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
		isc.NewEthereumAddressAgentID(from),
		isc.NewEthereumAddressAgentID(to),
		assets,
	)
}
