// Package defines interface to the persistent registry of the qnode
// The registry stores information about smart contracts and private keys and other data needed
// to sign the transaction
// all registry is cached in memory to enable fast check is SC transaction is of interest fo the node
// only SCMetaData records which node is processing is included in the cache
// if scid is not in cache, the transaction is ignored
package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/kvstore"
	. "github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/database"
	"io"
)

type SCMetaData struct {
	Address       address.Address
	Color         balance.Color
	OwnerAddress  address.Address
	Description   string
	ProgramHash   HashValue
	NodeLocations []string // "host_addr:port"
}

type SCMetaDataJsonable struct {
	Address       string   `json:"address"`       // base58
	Color         string   `json:"color"`         // base58
	OwnerAddress  string   `json:"owner_address"` // base58
	Description   string   `json:"description"`   // base58
	ProgramHash   string   `json:"program_hash"`  // base58
	NodeLocations []string `json:"node_locations"`
}

func (scd *SCMetaData) Jsonable() *SCMetaDataJsonable {
	return &SCMetaDataJsonable{
		Address:       scd.Address.String(),
		Color:         scd.Color.String(),
		OwnerAddress:  scd.OwnerAddress.String(),
		Description:   scd.Description,
		ProgramHash:   scd.ProgramHash.String(),
		NodeLocations: scd.NodeLocations,
	}
}

func (jo *SCMetaDataJsonable) ToSCMetaData() (*SCMetaData, error) {
	ret := &SCMetaData{
		Description:   jo.Description,
		NodeLocations: jo.NodeLocations,
	}
	var err error
	if ret.Address, err = address.FromBase58(jo.Address); err != nil {
		return nil, err
	}
	if ret.Color, err = util.ColorFromString(jo.Color); err != nil {
		return nil, err
	}
	if ret.OwnerAddress, err = address.FromBase58(jo.OwnerAddress); err != nil {
		return nil, err
	}
	if ret.ProgramHash, err = HashValueFromString(jo.ProgramHash); err != nil {
		return nil, err
	}
	return ret, nil
}

// GetScList retrieves all SCdata records from the registry
// in arbitrary key/value map order and returns a list
// if ownPortAddr is not nil, it only includes those SCMetaData records which are processed
// by his node
func GetSCDataList() ([]*SCMetaData, error) {
	db := database.GetRegistryPartition()
	ret := make([]*SCMetaData, 0)

	err := db.Iterate([]byte{database.ObjectTypeSCMetaData}, func(_ kvstore.Key, value kvstore.Value) bool {
		scdata := &SCMetaData{}
		if err := scdata.Read(bytes.NewReader(value)); err == nil {
			if validate(scdata) {
				ret = append(ret, scdata)
			}
		}
		return true
	})
	return ret, err
}

// checks if SCMetaData record is valid
// if ownAddr != nil checks if it is of interest to the current node
func validate(scdata *SCMetaData) bool {
	dkshare, ok, _ := GetDKShare(&scdata.Address)
	if !ok {
		// failed to load dkshare of the sc address
		return false
	}
	if int(dkshare.Index) >= len(scdata.NodeLocations) {
		// shouldn't be
		return false
	}
	return true
}

func dbkeyScdata(addr *address.Address) []byte {
	return database.MakeKey(database.ObjectTypeSCMetaData, addr[:])
}

// SaveSCData saves SCMetaData record to the registry
// overwrites previous if any for new sc
func SaveSCData(scd *SCMetaData) error {
	var buf bytes.Buffer
	if err := scd.Write(&buf); err != nil {
		return err
	}
	return database.GetRegistryPartition().Set(dbkeyScdata(&scd.Address), buf.Bytes())
}

func ExistSCMetaData(addr *address.Address) (bool, error) {
	return database.GetRegistryPartition().Has(dbkeyScdata(addr))
}

func GetSCData(addr *address.Address) (*SCMetaData, bool, error) {
	exists, err := ExistDKShareInRegistry(addr)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, fmt.Errorf("address is not known")
	}

	exists, err = ExistSCMetaData(addr)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}
	data, err := database.GetRegistryPartition().Get(dbkeyScdata(addr))
	ret := new(SCMetaData)
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, false, err
	}
	return ret, true, nil
}

func (scd *SCMetaData) Write(w io.Writer) error {
	if _, err := w.Write(scd.Address[:]); err != nil {
		return err
	}
	if _, err := w.Write(scd.Color[:]); err != nil {
		return err
	}
	if _, err := w.Write(scd.OwnerAddress[:]); err != nil {
		return err
	}
	if err := util.WriteString16(w, scd.Description); err != nil {
		return err
	}
	if _, err := w.Write(scd.ProgramHash[:]); err != nil {
		return err
	}
	if err := util.WriteStrings16(w, scd.NodeLocations); err != nil {
		return err
	}
	return nil
}

func (scd *SCMetaData) Read(r io.Reader) error {
	if err := util.ReadAddress(r, &scd.Address); err != nil {
		return err
	}
	if err := util.ReadColor(r, &scd.Color); err != nil {
		return err
	}
	if err := util.ReadAddress(r, &scd.OwnerAddress); err != nil {
		return err
	}
	var err error
	if scd.Description, err = util.ReadString16(r); err != nil {
		return err
	}
	if err := util.ReadHashValue(r, &scd.ProgramHash); err != nil {
		return err
	}
	if scd.NodeLocations, err = util.ReadStrings16(r); err != nil {
		return err
	}
	return nil
}
