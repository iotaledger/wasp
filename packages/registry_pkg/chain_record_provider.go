package registry_pkg

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/registry_pkg/chain_record"
	"github.com/mr-tron/base58"
)

// DB access
func MakeChainRecordDbKey(chainID *ledgerstate.AliasAddress) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, chainID.Bytes())
}

func (r *Impl) GetChainRecordByChainID(chainID *ledgerstate.AliasAddress) (*chain_record.ChainRecord, error) {
	data, err := r.store.Get(MakeChainRecordDbKey(chainID))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return chain_record.ChainRecordFromBytes(data)
}

func (r *Impl) GetChainRecords() ([]*chain_record.ChainRecord, error) {
	ret := make([]*chain_record.ChainRecord, 0)

	err := r.store.Iterate([]byte{dbkeys.ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		if rec, err1 := chain_record.ChainRecordFromBytes(value); err1 == nil {
			ret = append(ret, rec)
		} else {
			log.Warnf("corrupted chain record with key %s", base58.Encode(key))
		}
		return true
	})
	return ret, err
}

func (r *Impl) UpdateChainRecord(chainID *ledgerstate.AliasAddress, f func(*chain_record.ChainRecord) bool) (*chain_record.ChainRecord, error) {
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

func (r *Impl) ActivateChainRecord(chainID *ledgerstate.AliasAddress) (*chain_record.ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *chain_record.ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func (r *Impl) DeactivateChainRecord(chainID *ledgerstate.AliasAddress) (*chain_record.ChainRecord, error) {
	return r.UpdateChainRecord(chainID, func(bd *chain_record.ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func (r *Impl) SaveChainRecord(rec *chain_record.ChainRecord) error {
	key := dbkeys.MakeKey(dbkeys.ObjectTypeChainRecord, rec.ChainIdAliasAddress.Bytes())
	return r.store.Set(key, rec.Bytes())
}
