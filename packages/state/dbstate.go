package state

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
)

// database keys

const (
	stateUpdateDbPrefix      = "upd_"
	variableStateDbPrefix    = "vs_"
	requestProcessedDbPrefix = "rq_"
)

func stateUpdateStorageKey(addr *address.Address, stateIndex uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(stateUpdateDbPrefix))
	buf.Write(addr.Bytes())
	buf.Write(util.Uint32To4Bytes(stateIndex))
	return buf.Bytes()
}

func variableStateStorageKey(addr *address.Address) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(variableStateDbPrefix))
	buf.Write(addr.Bytes())
	return buf.Bytes()
}

func requestStorageKey(reqid *sctransaction.RequestId) []byte {
	var buf bytes.Buffer
	buf.Write([]byte(requestProcessedDbPrefix))
	buf.Write(reqid.Bytes())
	return buf.Bytes()
}

// loads state update with the given index
func LoadStateUpdate(addr *address.Address, stateIndex uint32) (StateUpdate, error) {
	storageKey := stateUpdateStorageKey(addr, stateIndex)
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return nil, err
	}
	exist, err := dbase.Contains(storageKey)
	if err != nil || !exist {
		return nil, err
	}
	entry, err := dbase.Get(storageKey)
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(entry.Value)
	ret := NewStateUpdate(nil, 0)
	if err = ret.Read(rdr); err != nil {
		return nil, err
	}
	// check consistency of the stored object
	if *ret.Address() != *addr || ret.StateIndex() != stateIndex {
		return nil, fmt.Errorf("LoadStateUpdate: invalid state update record in DB at state index %d address %s",
			stateIndex, addr.String())
	}
	return ret, nil
}

// saves state update to db
func (su *mockStateUpdate) SaveToDb() error {
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   stateUpdateStorageKey(su.address, su.stateIndex),
		Value: hashing.MustBytes(su),
	})
}

// loads variable state from db
func LoadVariableState(addr *address.Address) (VariableState, error) {
	storageKey := variableStateStorageKey(addr)
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return nil, err
	}
	exist, err := dbase.Contains(storageKey)
	if err != nil || !exist {
		return nil, err
	}
	entry, err := dbase.Get(storageKey)
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(entry.Value)
	ret := &mockVariableState{
		address:    addr,
		merkleHash: hashing.HashValue{},
	}
	if err = ret.Read(rdr); err != nil {
		return nil, err
	}
	return ret, nil
}

// saves variable state to db
func (vs *mockVariableState) SaveToDb() error {
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   variableStateStorageKey(vs.address),
		Value: hashing.MustBytes(vs),
	})
}

// marks request processed
// TODO time when processed, cleanup the index after some time and so on
func MarkRequestProcessed(reqid *sctransaction.RequestId) error {
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   requestStorageKey(reqid),
		Value: nil,
	})
}

// checks if request is processed
func IsRequestProcessed(reqid *sctransaction.RequestId) (bool, error) {
	storageKey := requestStorageKey(reqid)
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return false, err
	}
	exist, err := dbase.Contains(storageKey)
	if err != nil {
		return false, err
	}
	return exist, nil
}

// retrieves associated error string to the "request processed" record (if exists)
func RequestProcessedErrorString(reqid *sctransaction.RequestId) (string, error) {
	storageKey := requestStorageKey(reqid)
	dbase, err := database.GetSCStateDB()
	if err != nil {
		return "", err
	}
	entry, err := dbase.Get(storageKey)
	if err != nil {
		return "", err
	}
	return string(entry.Value), nil
}
