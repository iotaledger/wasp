// Package defines interface to the persistent registry of the qnode
// The registry stores information about smart contracts and private keys and other data needed
// to sign the transaction
// all registry is cached in memory to enable fast check is SC transaction is of interest fo the node
// only SCData records which node is processing is included in the cache
// if scid is not in cache, the transaction is ignored
package registry

import (
	"bytes"
	"encoding/json"
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

type SCData struct {
	Address       address.Address
	Color         balance.Color
	OwnerAddress  address.Address
	Description   string
	ProgramHash   HashValue
	NodeLocations []string // "host_addr:port"
}

// GetScList retrieves all SCdata records from the registry
// in arbitrary key/value map order and returns a list
// if ownPortAddr is not nil, it only includes those SCData records which are processed
// by his node
func GetSCDataList(ownAddr string) ([]*SCData, error) {
	dbase, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	ret := make([]*SCData, 0)
	err = dbase.ForEachPrefix(dbSCDataGroupPrefix, func(entry database.Entry) bool {
		scdata := &SCData{}
		if err = json.Unmarshal(entry.Value, scdata); err == nil {
			if validate(scdata, ownAddr) {
				ret = append(ret, scdata)
			}
		}
		return false
	})
	return ret, err
}

// checks if SCData record is valid
// if ownAddr != nil checks if it is of interest to the current node
func validate(scdata *SCData, ownAddr string) bool {
	dkshare, ok, _ := GetDKShare(&scdata.Address)
	if !ok {
		// failed to load dkshare of the sc address
		return false
	}
	if int(dkshare.Index) >= len(scdata.NodeLocations) {
		// shouldn't be
		return false
	}
	if ownAddr == "" {
		return true
	}
	if ownAddr != scdata.NodeLocations[dkshare.Index] {
		// if own address is not consistent with the one at the index in the list of nodes
		// this node is not interested in the SC
		return false
	}
	return true
}

// prefix of the SCData entry key
var dbSCDataGroupPrefix = HashStrings("scdata").Bytes()

// key of the entry
func dbSCDataKey(addr *address.Address) []byte {
	var buf bytes.Buffer
	buf.Write(dbSCDataGroupPrefix)
	buf.Write(addr.Bytes())
	return buf.Bytes()
}

// SaveSCData saves SCData record to the registry
// overwrites previous if any
// for new sc
func SaveSCData(scd *SCData) error {
	dbase, err := database.GetDB()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(scd)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbSCDataKey(&scd.Address),
		Value: jsonData,
	})
}

func GetSCData(addr *address.Address) (*SCData, error) {
	dbase, err := database.GetDB()
	if err != nil {
		return nil, err
	}
	entry, err := dbase.Get(dbSCDataKey(addr))
	if err != nil {
		return nil, err
	}
	var ret SCData
	if err := json.Unmarshal(entry.Value, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (scd *SCData) Write(w io.Writer) error {
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

func (scd *SCData) Read(r io.Reader) error {
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
