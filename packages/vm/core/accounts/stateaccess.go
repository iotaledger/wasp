// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accounts

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
	return codec.AgentID.Decode([]byte(key))
}
