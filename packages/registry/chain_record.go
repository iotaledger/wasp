package registry

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/mr-tron/base58"
)

// ChainRecord represents chain the node is participating in
// TODO optimize, no need for a persistent structure, simple activity tag is enough
type ChainRecord struct {
	ChainID *coretypes.ChainID
	Active  bool
}

func NewChainRecord(chainID *coretypes.ChainID, active ...bool) *ChainRecord {
	act := false
	if len(active) > 0 {
		act = active[0]
	}
	return &ChainRecord{
		ChainID: chainID.Clone(),
		Active:  act,
	}
}

func ChainRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ChainRecord, error) {
	ret := &ChainRecord{}
	aliasAddr, err := ledgerstate.AliasAddressFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	ret.ChainID = coretypes.NewChainID(aliasAddr)

	ret.Active, err = mu.ReadBool()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// CommitteeRecordFromBytes
func ChainRecordFromBytes(data []byte) (*ChainRecord, error) {
	return ChainRecordFromMarshalUtil(marshalutil.New(data))
}

// ChainRecordFromRegistry reads ChainRecord from registry.
// Returns nil if not found
func ChainRecordFromRegistry(chainID *coretypes.ChainID) (*ChainRecord, error) {
	data, err := database.GetRegistryPartition().Get(dbKeyChainRecord(chainID))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ChainRecordFromBytes(data)
}

func (rec *ChainRecord) Bytes() []byte {
	return marshalutil.New().
		WriteBytes(rec.ChainID.Bytes()).
		WriteBool(rec.Active).
		Bytes()
}

func dbKeyChainRecord(chainID *coretypes.ChainID) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeChainRecord, chainID.Bytes())
}

func (rec *ChainRecord) SaveToRegistry() error {
	return database.GetRegistryPartition().Set(dbKeyChainRecord(rec.ChainID), rec.Bytes())
}

func UpdateChainRecord(chainID *coretypes.ChainID, f func(*ChainRecord) bool) (*ChainRecord, error) {
	rec, err := ChainRecordFromRegistry(chainID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("no chain record found for chainID %s", chainID.String())
	}
	if f(rec) {
		err = rec.SaveToRegistry()
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func ActivateChainRecord(chainID *coretypes.ChainID) (*ChainRecord, error) {
	return UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func DeactivateChainRecord(chainID *coretypes.ChainID) (*ChainRecord, error) {
	return UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
}

func GetChainRecords() ([]*ChainRecord, error) {
	db := database.GetRegistryPartition()
	ret := make([]*ChainRecord, 0)

	err := db.Iterate([]byte{dbprovider.ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		if rec, err1 := ChainRecordFromBytes(value); err1 == nil {
			ret = append(ret, rec)
		} else {
			log.Warnf("corrupted chain record with key %s", base58.Encode(key))
		}
		return true
	})
	return ret, err
}

func (rec *ChainRecord) String() string {
	ret := "ChainID: " + rec.ChainID.String() + "\n"
	ret += fmt.Sprintf("      Active: %v\n", rec.Active)
	return ret
}
