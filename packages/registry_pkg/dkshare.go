// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry_pkg

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// SaveDKShare implements dkg.DKShareRegistryProvider.
func (r *Impl) SaveDKShare(dkShare *tcrypto.DKShare) error {
	var err error
	var exists bool
	dbKey := dbKeyForDKShare(dkShare.Address)

	if exists, err = r.store.Has(dbKey); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}
	var buf []byte
	if buf, err = dkShare.Bytes(); err != nil {
		return err
	}
	return r.store.Set(dbKey, buf)

}

// LoadDKShare implements dkg.DKShareRegistryProvider.
func (r *Impl) LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error) {
	data, err := r.store.Get(dbKeyForDKShare(sharedAddress))
	if err != nil {
		return nil, err
	}
	return tcrypto.DKShareFromBytes(data, r.suite)
}

func dbKeyForDKShare(sharedAddress ledgerstate.Address) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeDistributedKeyData, sharedAddress.Bytes())
}
