// Package defines interface to the persistent registry of the qnode
// The registry stores information about smart contracts and private keys and other data needed
// to sign the transaction
// all registry is cached in memory to enable fast check is SC transaction is of interest fo the node
// only SCMetaData records which node is processing is included in the cache
// if scid is not in cache, the transaction is ignored
package registry

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/database"
	. "github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

// the data structure introduces smart contract to wasp
// the color is origin transaction hash
// TODO add state update variables

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

// GetScList retrieves all SCdata records from the registry
// in arbitrary key/value map order and returns a list
// if ownPortAddr is not nil, it only includes those SCMetaData records which are processed
// by his node
func GetSCDataList() ([]*SCMetaData, error) {
	dbase, err := database.GetSCMetaDataDB()
	if err != nil {
		return nil, err
	}
	ret := make([]*SCMetaData, 0)
	err = dbase.ForEachPrefix(nil, func(entry database.Entry) bool {
		scdata := &SCMetaData{}
		if err = scdata.Read(bytes.NewReader(entry.Value)); err == nil {
			if validate(scdata) {
				ret = append(ret, scdata)
			}
		}
		return false
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

// SaveSCData saves SCMetaData record to the registry
// overwrites previous if any
// for new sc
func SaveSCData(scd *SCMetaData) error {
	dbase, err := database.GetSCMetaDataDB()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := scd.Write(&buf); err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   database.DbKeySCMetaData(&scd.Address),
		Value: buf.Bytes(),
	})
}

func GetSCData(addr *address.Address) (*SCMetaData, error) {
	dbase, err := database.GetSCMetaDataDB()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(database.DbKeySCMetaData(addr))
	if err != nil {
		return nil, err
	}
	var ret SCMetaData
	if err := ret.Read(bytes.NewReader(entry.Value)); err != nil {
		return nil, err
	}
	return &ret, nil
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
