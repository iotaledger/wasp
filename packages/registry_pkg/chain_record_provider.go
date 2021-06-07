package registry_pkg

import (
	"fmt"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"
	"github.com/mr-tron/base58"
)

// DB access
func MakeChainRecordDbKey(chainID *coretypes.ChainID) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, chainID.Bytes())
}

func (r *Impl) GetChainRecordByChainID(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error) {
	data, err := r.store.Get(MakeChainRecordDbKey(chainID))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return chainrecord.ChainRecordFromBytes(data)
}

func (r *Impl) GetChainRecords() ([]*chainrecord.ChainRecord, error) {
	ret := make([]*chainrecord.ChainRecord, 0)

	err := r.store.Iterate([]byte{dbkeys.ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		if rec, err1 := chainrecord.ChainRecordFromBytes(value); err1 == nil {
			ret = append(ret, rec)
		} else {
			log.Warnf("corrupted chain record with key %s", base58.Encode(key))
		}
		return true
	})
	return ret, err
}

func (r *Impl) UpdateChainRecord(chainID *coretypes.ChainID, f func(*chainrecord.ChainRecord) bool) (*chainrecord.ChainRecord, error) {
	rec, err := r.GetChainRecordByChainID(chainID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("no chain record found for chainID %s", chainID.String())
	}
	if f(rec) {
		err = r.SaveChainRecord(rec)
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func (r *Impl) ActivateChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *chainrecord.ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func (r *Impl) DeactivateChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *chainrecord.ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func (r *Impl) SaveChainRecord(rec *chainrecord.ChainRecord) error {
	key := dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, rec.ChainID.Bytes())
	return r.store.Set(key, rec.Bytes())
}
