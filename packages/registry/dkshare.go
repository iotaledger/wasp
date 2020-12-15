// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/database"
)

// SaveDKShare implements dkg.RegistryProvider.
func (r *Impl) SaveDKShare(dkShare *tcrypto.DKShare) error {
	var err error
	var exists bool
	dbKey := dbKeyForDKShare(dkShare.Address)
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
func (r *Impl) LoadDKShare(sharedAddress *address.Address) (*tcrypto.DKShare, error) {
	data, err := database.GetRegistryPartition().Get(dbKeyForDKShare(sharedAddress))
	if err != nil {
		return nil, err
	}
	return tcrypto.DKShareFromBytes(data, r.suite)
}

func dbKeyForDKShare(sharedAddress *address.Address) []byte {
	return database.MakeKey(database.ObjectTypeDistributedKeyData, sharedAddress.Bytes())
}
