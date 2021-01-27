package registry

import (
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/database"
)

func dbKeyForBlob(h hashing.HashValue) []byte {
	return dbprovider.MakeKey(dbprovider.ObjectTypeBlob, h[:])
}

// Writes data into the registry with the key of its hash
func (r *Impl) PutBlob(data []byte) error {
	h := hashing.HashData(data)
	return database.GetRegistryPartition().Set(h[:], data)
}

// Reads data from registry by hash. Returns existence flag
func (r *Impl) GetBlob(h hashing.HashValue) ([]byte, bool, error) {
	ret, err := database.GetRegistryPartition().Get(dbKeyForBlob(h))
	return ret, ret != nil && err == nil, err
}

// DecodeRequestArguments decodes RequestArguments. For each blog reference encoded it
// looks for the data by hash into the registry and replaces dict entry with the data
// It returns ok flag == false if at least one blob hash don't have data in the registry
func (r *Impl) DecodeRequestArguments(reqArgs sctransaction.RequestArguments) (dict.Dict, bool, error) {
	ret := dict.New()
	reqArgsDict := dict.Dict(reqArgs)
	var ok bool
	var err error
	var data []byte
	var h hashing.HashValue
	reqArgsDict.ForEach(func(key kv.Key, value []byte) bool {
		d := []byte(key)
		if d[0] == '*' {
			h, err = hashing.HashValueFromBytes(value)
			if err != nil {
				return false
			}
			data, ok, err = r.GetBlob(h)
			if err != nil {
				return false
			}
			if !ok {
				return false
			}
			ret.Set(kv.Key(d[1:]), data)
		} else {
			ret.Set(kv.Key(d[1:]), value)
		}
		return true
	})
	if err != nil || !ok {
		ret = nil
	}
	return ret, ok, err
}
