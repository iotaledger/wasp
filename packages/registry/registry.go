// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	suite tcrypto.Suite
	log   *logger.Logger
	store kvstore.KVStore
}

// New creates new instance of the registry implementation.
func NewRegistry(suite tcrypto.Suite, log *logger.Logger, store kvstore.KVStore) *Impl {
	if store == nil {
		store = database.GetRegistryKVStore()
	}
	ret := &Impl{
		suite: suite,
		log:   log.Named("registry"),
		store: store,
	}

	return ret
}

func dbKeyCommitteeRecord(addr ledgerstate.Address) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeCommitteeRecord, addr.Bytes())
}

func (r *Impl) GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	data, err := r.store.Get(dbKeyCommitteeRecord(addr))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return committee_record.CommitteeRecordFromBytes(data)
}

func (r *Impl) SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error {
	return r.store.Set(dbKeyCommitteeRecord(rec.Address), rec.Bytes())
}
