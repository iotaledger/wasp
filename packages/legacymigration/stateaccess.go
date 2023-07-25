// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package legacymigration

import (
	"github.com/iotaledger/hive.go/lo"
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

func (sa *StateAccess) MigratableBalance(legacyAddr []byte) uint64 {
	return migratableBalance(sa.state, legacyAddr)
}

func (sa *StateAccess) ValidMigrationRequest(req isc.Request) bool {
	if req.CallTarget().Contract != Contract.Hname() { // must call this contract
		return false
	}

	if req.CallTarget().EntryPoint != FuncMigrate.Hname() { // must call migration entrypoint
		return false
	}

	bundleBytes := req.Params().Get(ParamBundle)
	bndl, err := validBundleFromBytes(bundleBytes)
	if err != nil {
		return false
	}
	migratedAddresses, targetAddress, err := addressesFromBundle(bndl)
	if err != nil || len(migratedAddresses) == 0 || targetAddress == nil {
		return false
	}
	fundsToMigrate := lo.Reduce(migratedAddresses, func(acc uint64, legacyAddr []byte) uint64 {
		return acc + sa.MigratableBalance(legacyAddr)
	}, uint64(0))
	return fundsToMigrate > 0
}
