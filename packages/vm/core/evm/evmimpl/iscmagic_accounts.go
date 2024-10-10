// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCAccounts::getL2BalanceBaseTokens
func (h *magicContractHandler) GetL2BalanceBaseTokens(agentID iscmagic.ISCAgentID) coin.Value {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewBalanceBaseToken.Message(&aid))
	return lo.Must(accounts.ViewBalanceBaseToken.DecodeOutput(r))
}

// handler for ISCAccounts::getL2BalanceCoin
func (h *magicContractHandler) GetL2BalanceCoin(
	coinType coin.Type,
	agentID iscmagic.ISCAgentID,
) coin.Value {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewBalanceCoin.Message(&aid, coinType))
	return lo.Must(accounts.ViewBalanceCoin.DecodeOutput(r))
}

// handler for ISCAccounts::getL2Objects
func (h *magicContractHandler) GetL2Objects(agentID iscmagic.ISCAgentID) []sui.ObjectID {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewAccountObjects.Message(&aid))
	return lo.Must(accounts.ViewAccountObjects.DecodeOutput(r))
}

// handler for ISCAccounts::getL2ObjectCount
func (h *magicContractHandler) GetL2ObjectCount(agentID iscmagic.ISCAgentID) *big.Int {
	// TODO: avoid fetching the whole list of objects
	return new(big.Int).SetUint64(uint64(len(h.GetL2Objects(agentID))))
}

// handler for ISCAccounts::getL2ObjectsInCollection
func (h *magicContractHandler) GetL2ObjectsInCollection(
	agentID iscmagic.ISCAgentID,
	collectionID sui.ObjectID,
) []sui.ObjectID {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewAccountObjectsInCollection.Message(&aid, collectionID))
	return lo.Must(accounts.ViewAccountObjectsInCollection.DecodeOutput(r))
}

// handler for ISCAccounts::getL2ObjectsCountInCollection
func (h *magicContractHandler) GetL2ObjectsCountInCollection(
	agentID iscmagic.ISCAgentID,
	collectionID sui.ObjectID,
) *big.Int {
	// TODO: avoid fetching the whole list of objects
	return new(big.Int).SetUint64(uint64(len(h.GetL2ObjectsInCollection(agentID, collectionID))))
}
