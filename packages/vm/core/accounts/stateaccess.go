// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accounts

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
)

type StateAccess struct {
	chainState state.State
	state      kv.KVStoreReader
}

func NewStateAccess(chainState state.State) *StateAccess {
	state := subrealm.NewReadOnly(chainState, kv.Key(Contract.Hname().Bytes()))
	return &StateAccess{state: state, chainState: chainState}
}

func (sa *StateAccess) Nonce(agentID isc.AgentID, chainID isc.ChainID) uint64 {
	return AccountNonce(sa.state, agentID, chainID)
}

func (sa *StateAccess) AccountExists(agentID isc.AgentID, chainID isc.ChainID) bool {
	return accountExists(sa.state, agentID, chainID)
}

// converts an account key from the accounts contract (shortform without chainID) to an AgentID
func AgentIDFromKey(key kv.Key, chainID isc.ChainID) (isc.AgentID, error) {
	if len(key) < isc.ChainIDLength {
		// short form saved (withoutChainID)
		switch len(key) {
		case 4:
			hn, err := isc.HnameFromBytes([]byte(key))
			if err != nil {
				return nil, err
			}
			return isc.NewContractAgentID(chainID, hn), nil
		case common.AddressLength:
			var ethAddr common.Address
			copy(ethAddr[:], []byte(key))
			return isc.NewEthereumAddressAgentID(chainID, ethAddr), nil
		default:
			panic("bad key length")
		}
	}
	return codec.DecodeAgentID([]byte(key))
}

func (sa *StateAccess) AllAccounts() dict.Dict {
	return AllAccountsAsDict(sa.state)
}

func (sa *StateAccess) IterateAccounts() func(func(key []byte) bool) {
	return AllAccountsMapR(sa.state).IterateKeys
}

// NOTE: passing the AgentID seems silly, but it's necessary because NFT's don't follow the same logic as the fungible tokens, and are instead stored by full agentID
func (sa *StateAccess) AssetsOwnedBy(accKey kv.Key, agentID isc.AgentID) *isc.Assets {
	ret := getFungibleTokens(sa.chainState.SchemaVersion(), sa.state, accKey)
	ret.AddNFTs(getAccountNFTs(sa.state, agentID)...)
	return ret
}
