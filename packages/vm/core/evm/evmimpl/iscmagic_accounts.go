// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
)

// handler for ISCAccounts::getL2BalanceBaseTokens
func (h *magicContractHandler) GetL2BalanceBaseTokens(agentID iscmagic.ISCAgentID) iscmagic.CoinValue {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewBalanceBaseToken.Message(&aid))
	return iscmagic.CoinValue(lo.Must(accounts.ViewBalanceBaseToken.DecodeOutput(r)))
}

// handler for ISCAccounts::getL2BalanceCoin
func (h *magicContractHandler) GetL2BalanceCoin(
	coinType iscmagic.CoinType,
	agentID iscmagic.ISCAgentID,
) coin.Value {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewBalanceCoin.Message(&aid, coin.MustTypeFromString(coinType)))
	return lo.Must(accounts.ViewBalanceCoin.DecodeOutput(r))
}

// handler for ISCAccounts::getL2Objects
func (h *magicContractHandler) GetL2Objects(agentID iscmagic.ISCAgentID) []iotago.ObjectID {
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
	collectionID iotago.ObjectID,
) []iotago.ObjectID {
	aid := lo.Must(agentID.Unwrap())
	r := h.callView(accounts.ViewAccountObjectsInCollection.Message(&aid, collectionID))
	return lo.Must(accounts.ViewAccountObjectsInCollection.DecodeOutput(r))
}

// handler for ISCAccounts::getL2ObjectsCountInCollection
func (h *magicContractHandler) GetL2ObjectsCountInCollection(
	agentID iscmagic.ISCAgentID,
	collectionID iotago.ObjectID,
) *big.Int {
	// TODO: avoid fetching the whole list of objects
	return new(big.Int).SetUint64(uint64(len(h.GetL2ObjectsInCollection(agentID, collectionID))))
}
