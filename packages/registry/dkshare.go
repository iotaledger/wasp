package registry

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dks"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/database"
	"go.dedis.ch/kyber/v3"
)

func CommitDKShare(ks *tcrypto.DKShare, pubKeys []kyber.Point) error {
	if err := ks.FinalizeDKS(pubKeys); err != nil {
		return err
	}
	return SaveDKShareToRegistry(ks)
}

func dbkey(addr *address.Address) []byte {
	return database.MakeKey(database.ObjectTypeDistributedKeyData, addr.Bytes())
}

func SaveDKShareToRegistry(ks *tcrypto.DKShare) error {
	if !ks.Committed {
		return fmt.Errorf("uncommited DK share: can't be saved to the registry")
	}
	dbase := database.GetRegistryPartition()
	exists, err := dbase.Has(dbkey(ks.Address))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}

	var buf bytes.Buffer

	err = ks.Write(&buf)
	if err != nil {
		return err
	}
	return dbase.Set(dbkey(ks.Address), buf.Bytes())
}

func GetDKShare(addr *address.Address) (*tcrypto.DKShare, bool, error) {
	exist, err := database.GetRegistryPartition().Has(dbkey(addr))

	if err != nil || !exist {
		return nil, false, err
	}
	ret, err := LoadDKShare(addr, false)
	if err != nil {
		return nil, false, err
	}
	return ret, true, nil
}

func LoadDKShare(addr *address.Address, maskPrivate bool) (*tcrypto.DKShare, error) {
	data, err := database.GetRegistryPartition().Get(dbkey(addr))
	if err != nil {
		return nil, err
	}
	ret, err := tcrypto.UnmarshalDKShare(data, maskPrivate)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

//
// TODO: It should be used instead of the above after migration is complete.
//

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
	return dks.DKShareFromBytes(data, r.groupSuite)
}

func dbKeyForDKShare(chainID *coretypes.ChainID) []byte {
	return database.MakeKey(database.ObjectTypeDistributedKeyData, chainID.Bytes())
}
