// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

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
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), sender),
		isc.NewEthereumAddressAgentID(h.ctx.ChainID(), receiver),
		allowance.Unwrap(),
	)
}

// handler for ISCPrivileged::addToAllowance
// Deprecated: called from previous versions of ERC20NativeTokens.sol and
// ERC20BaseTokens.sol. May be removed after all living chains are spawned
// with version > v0.6.1-alpha.12.
func (h *magicContractHandler) AddToAllowance(
	from common.Address,
	to common.Address,
	allowance iscmagic.ISCAssets,
) {
	addToAllowance(h.ctx, from, to, allowance.Unwrap())
}

// handler for ISCPrivileged::setAllowanceBaseTokens
func (h *magicContractHandler) SetAllowanceBaseTokens(
	from common.Address,
	to common.Address,
	numTokens *big.Int,
) {
	setAllowanceBaseTokens(h.ctx, from, to, numTokens)
}

// handler for ISCPrivileged::setAllowanceNativeTokens
func (h *magicContractHandler) SetAllowanceNativeTokens(
	from common.Address,
	to common.Address,
	nativeTokenID iscmagic.NativeTokenID,
	numTokens *big.Int,
) {
	setAllowanceNativeTokens(h.ctx, from, to, nativeTokenID, numTokens)
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
