package registry

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dks"
	"github.com/iotaledger/wasp/plugins/database"
)

// SaveDKShare implements dkg.RegistryProvider.
func (r *Impl) SaveDKShare(dkShare *dks.DKShare) error {
	var err error
	var exists bool
	dbKey := dbKeyForDKShare(dkShare.ChainID)
	kvStore := database.GetRegistryPartition()
	if exists, err = kvStore.Has(dbKey); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}
	var buf []byte
	if buf, err = dkShare.Bytes(); err != nil {
		return err
	}
	return kvStore.Set(dbKey, buf)

}

// LoadDKShare implements dkg.RegistryProvider.
func (r *Impl) LoadDKShare(chainID *coretypes.ChainID) (*dks.DKShare, error) {
	data, err := database.GetRegistryPartition().Get(dbKeyForDKShare(chainID))
	if err != nil {
		return nil, err
	}
	return dks.DKShareFromBytes(data, r.suite)
}

func dbKeyForDKShare(chainID *coretypes.ChainID) []byte {
	return database.MakeKey(database.ObjectTypeDistributedKeyData, chainID.Bytes())
}
