package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"github.com/iotaledger/wasp/plugins/publisher"
	"github.com/mr-tron/base58"
	"io"
)

// BootupData is a minimum data needed to load a committee for the smart contract
// it is up to the node (not smart contract) to check authorisations to create/update this record
type BootupData struct {
	Address        address.Address
	OwnerAddress   address.Address // only needed for committee nodes, can be nil for access nodes
	Color          balance.Color   // origin tx hash
	CommitteeNodes []string        // "host_addr:port"
	AccessNodes    []string        // "host_addr:port"
	Active         bool
}

func dbkeyBootupData(addr *address.Address) []byte {
	return database.MakeKey(database.ObjectTypeBootupData, addr[:])
}

func SaveBootupData(bd *BootupData) error {
	var niladdr address.Address
	if bd.Address == niladdr {
		return fmt.Errorf("can be empty address")
	}
	if bd.Color == balance.ColorNew || bd.Color == balance.ColorIOTA {
		return fmt.Errorf("can't be IOTA or New color")
	}
	var buf bytes.Buffer
	if err := bd.Write(&buf); err != nil {
		return err
	}
	if err := database.GetRegistryPartition().Set(dbkeyBootupData(&bd.Address), buf.Bytes()); err != nil {
		return err
	}
	publisher.Publish("bootuprec", bd.Address.String(), bd.Color.String())
	return nil
}

func GetBootupData(addr *address.Address) (*BootupData, error) {
	data, err := database.GetRegistryPartition().Get(dbkeyBootupData(addr))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ret := new(BootupData)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateBootupData(addr *address.Address, f func(*BootupData) bool) (*BootupData, error) {
	bd, err := GetBootupData(addr)
	if err != nil {
		return nil, err
	}
	if bd == nil {
		return nil, fmt.Errorf("No bootup data found for address %s", addr)
	}
	if f(bd) {
		err = SaveBootupData(bd)
		if err != nil {
			return nil, err
		}
	}
	return bd, nil
}

func ActivateBootupData(addr *address.Address) (*BootupData, error) {
	return UpdateBootupData(addr, func(bd *BootupData) bool {
		if bd.Active {
			return false
		}
		bd.Active = true
		return true
	})
}

func DeactivateBootupData(addr *address.Address) (*BootupData, error) {
	return UpdateBootupData(addr, func(bd *BootupData) bool {
		if !bd.Active {
			return false
		}
		bd.Active = false
		return true
	})
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
	if _, err := w.Write(bd.OwnerAddress[:]); err != nil {
		return err
	}
	if _, err := w.Write(bd.Color[:]); err != nil {
		return err
	}
	if err := util.WriteStrings16(w, bd.CommitteeNodes); err != nil {
		return err
	}
	if err := util.WriteStrings16(w, bd.AccessNodes); err != nil {
		return err
	}
	if err := util.WriteBoolByte(w, bd.Active); err != nil {
		return err
	}
	return nil
}

func (bd *BootupData) Read(r io.Reader) error {
	var err error
	if err = util.ReadAddress(r, &bd.Address); err != nil {
		return err
	}
	if err = util.ReadAddress(r, &bd.OwnerAddress); err != nil {
		return err
	}
	if err = util.ReadColor(r, &bd.Color); err != nil {
		return err
	}
	if bd.CommitteeNodes, err = util.ReadStrings16(r); err != nil {
		return err
	}
	if bd.AccessNodes, err = util.ReadStrings16(r); err != nil {
		return err
	}
	if err = util.ReadBoolByte(r, &bd.Active); err != nil {
		return err
	}
	return nil
}

func (bd *BootupData) String() string {
	ret := "      Address: " + bd.Address.String() + "\n"
	ret += "      Color: " + bd.Color.String() + "\n"
	ret += "      Owner address: " + bd.OwnerAddress.String() + "\n"
	ret += fmt.Sprintf("      Committee nodes: %+v\n", bd.CommitteeNodes)
	ret += fmt.Sprintf("      Access nodes: %+v\n", bd.AccessNodes)
	return ret
}
