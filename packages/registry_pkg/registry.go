// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry_pkg

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/registry_pkg/committee_record"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/database"
)

// Impl is just a placeholder to implement all interfaces needed by different components.
// Each of the interfaces are implemented in the corresponding file in this package.
type Impl struct {
	suite      tcrypto.Suite
	log        *logger.Logger
	dbProvider *dbprovider.DBProvider
}

// New creates new instance of the registry implementation.
func NewRegistry(suite tcrypto.Suite, log *logger.Logger, dbp ...*dbprovider.DBProvider) *Impl {
	ret := &Impl{
		suite: suite,
		log:   log.Named("registry"),
	}
	if len(dbp) == 0 {
		ret.dbProvider = database.GetInstance()
	} else {
		ret.dbProvider = dbp[0]
	}
	return ret
}

func (i *Impl) DBProvider() *dbprovider.DBProvider {
	return i.dbProvider
}

func dbKeyCommitteeRecord(addr ledgerstate.Address) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeCommitteeRecord, addr.Bytes())
}

func (r *Impl) GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	data, err := r.dbProvider.GetRegistryPartition().Get(dbKeyCommitteeRecord(addr))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return committee_record.CommitteeRecordFromBytes(data)
}

func (r *Impl) SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error {
	return r.dbProvider.GetRegistryPartition().Set(dbKeyCommitteeRecord(rec.Address), rec.Bytes())
}
