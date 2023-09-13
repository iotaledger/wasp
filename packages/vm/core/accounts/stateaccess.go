// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

type StateAccess struct {
	state kv.KVStoreReader
}

func NewStateAccess(store kv.KVStoreReader) *StateAccess {
	state := subrealm.NewReadOnly(store, kv.Key(Contract.Hname().Bytes()))
	return &StateAccess{state: state}
}

func (sa *StateAccess) Nonce(agentID isc.AgentID, chainID isc.ChainID) uint64 {
	return accountNonce(sa.state, agentID, chainID)
}

func (sa *StateAccess) AccountExists(agentID isc.AgentID, chainID isc.ChainID) bool {
	return accountExists(sa.state, agentID, chainID)
}
