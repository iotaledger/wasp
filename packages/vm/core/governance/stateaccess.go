// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
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

func (sa *StateAccess) GetMaintenanceStatus() bool {
	r := sa.state.Get(VarMaintenanceStatus)
	if r == nil {
		return false // chain is being initialized, governance has not been initialized yet
	}
	return codec.MustDecodeBool(r)
}

func (sa *StateAccess) GetAccessNodes() []*cryptolib.PublicKey {
	accessNodes := []*cryptolib.PublicKey{}
	AccessNodesMapR(sa.state).IterateKeys(func(pubKeyBytes []byte) bool {
		pubKey, err := cryptolib.NewPublicKeyFromBytes(pubKeyBytes)
		if err != nil {
			panic(err)
		}
		accessNodes = append(accessNodes, pubKey)
		return true
	})
	return accessNodes
}

func (sa *StateAccess) GetCandidateNodes() []*AccessNodeInfo {
	candidateNodes := []*AccessNodeInfo{}
	AccessNodeCandidatesMapR(sa.state).Iterate(func(pubKeyBytes, accessNodeInfoBytes []byte) bool {
		ani, err := NewAccessNodeInfoFromBytes(pubKeyBytes, accessNodeInfoBytes)
		if err != nil {
			panic(err)
		}
		candidateNodes = append(candidateNodes, ani)
		return true
	})
	return candidateNodes
}
