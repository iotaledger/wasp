package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/mr-tron/base58"
	"io"
)

// BootupData is a minimum data needed to load a committee for the smart contract
// it is up to the node (not smart contract) to check authorisations to create/update this record
type BootupData struct {
	Address       address.Address
	NodeLocations []string // "host_addr:port"
}

func dbkeyBootupData(addr *address.Address) []byte {
	return database.MakeKey(database.ObjectTypeBootupData, addr[:])
}

func SaveBootupData(bd *BootupData, overwrite bool) error {
	if overwrite {
		exist, err := database.GetRegistryPartition().Has(dbkeyBootupData(&bd.Address))
		if err != nil {
			return err
		}
		if exist {
			return fmt.Errorf("smart contract with address %s aldready exist in the registry", bd.Address.String())
		}
	}
	var buf bytes.Buffer
	if err := bd.Write(&buf); err != nil {
		return err
	}
	return database.GetRegistryPartition().Set(dbkeyBootupData(&bd.Address), buf.Bytes())
}

func GetBootupData(addr *address.Address) (*BootupData, bool, error) {
	data, err := database.GetRegistryPartition().Get(dbkeyBootupData(addr))
	if err == kvstore.ErrKeyNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	ret := new(BootupData)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, false, err
	}
	return ret, true, nil
}

func GetBootupRecords() ([]*BootupData, error) {
	db := database.GetRegistryPartition()
	ret := make([]*BootupData, 0)

	err := db.Iterate([]byte{database.ObjectTypeBootupData}, func(key kvstore.Key, value kvstore.Value) bool {
		bd := new(BootupData)
		if err := bd.Read(bytes.NewReader(value)); err == nil {
			ret = append(ret, bd)
		} else {
			log.Warnf("corrupted bootup record with key %s", base58.Encode(key))
		}
		return true
	})
	return ret, err
}

func (bd *BootupData) Write(w io.Writer) error {
	if _, err := w.Write(bd.Address[:]); err != nil {
		return err
	}
	if err := util.WriteStrings16(w, bd.NodeLocations); err != nil {
		return err
	}
	return nil
}

func (bd *BootupData) Read(r io.Reader) error {
	var err error
	if err = util.ReadAddress(r, &bd.Address); err != nil {
		return err
	}

	if bd.NodeLocations, err = util.ReadStrings16(r); err != nil {
		return err
	}
	return nil
}
