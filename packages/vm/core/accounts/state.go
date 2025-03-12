// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accounts

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type StateReader struct {
	v     isc.SchemaVersion
	state kv.KVStoreReader
}

func NewStateReader(v isc.SchemaVersion, contractState kv.KVStoreReader) *StateReader {
	return &StateReader{
		v:     v,
		state: contractState,
	}
}

func NewStateReaderFromSandbox(ctx isc.SandboxBase) *StateReader {
	return NewStateReader(ctx.SchemaVersion(), ctx.StateR())
}

func NewStateReaderFromChainState(v isc.SchemaVersion, chainState kv.KVStoreReader) *StateReader {
	return NewStateReader(v, Contract.StateSubrealmR(chainState))
}

type StateWriter struct {
	*StateReader
	state kv.KVStore
}

func NewStateWriter(v isc.SchemaVersion, contractState kv.KVStore) *StateWriter {
	return &StateWriter{
		StateReader: NewStateReader(v, contractState),
		state:       contractState,
	}
}

func NewStateWriterFromSandbox(ctx isc.Sandbox) *StateWriter {
	return NewStateWriter(ctx.SchemaVersion(), ctx.State())
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
			panic(fmt.Sprintf("bad key length: %d: %v / %x", len(key), string(key), key))
		}
	}
	return codec.Decode[isc.AgentID]([]byte(key))
}
