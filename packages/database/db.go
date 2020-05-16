package database

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/database"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
)

const (
	DBPrefixDatabaseVersion byte = iota
	DBPrefixSCMetaData
	DBPrefixKeyData
	DBPrefixSCVariableState
	DBPrefixSCStateUpdates
	DBPrefixProcessedRequests
)

// smart contract meta data
func GetSCMetaDataDB() (database.Database, error) {
	return Get(DBPrefixSCMetaData, GetBadgerInstance())
}

// smart contract meta data
func GetKeyDataDB() (database.Database, error) {
	return Get(DBPrefixKeyData, GetBadgerInstance())
}

// smart contract state data
func GetVariableStateDB() (database.Database, error) {
	return Get(DBPrefixSCVariableState, GetBadgerInstance())
}

// smart contract state update data
func GetStateUpdateDB() (database.Database, error) {
	return Get(DBPrefixSCStateUpdates, GetBadgerInstance())
}

// smart contract state update data
func GetProcessedRequestsDB() (database.Database, error) {
	return Get(DBPrefixProcessedRequests, GetBadgerInstance())
}

// omitting first version byte in the address
func ObjKey(fragments ...[]byte) []byte {
	var buf bytes.Buffer
	for _, f := range fragments {
		buf.Write(f)
	}
	return buf.Bytes()
}

// omitting first version byte in the address
func ObjAddressKey(addr *address.Address, fragments ...[]byte) []byte {
	return ObjKey(append([][]byte{addr.Bytes()}, fragments...)...)
}

func DbKeyDKShare(addr *address.Address) []byte {
	return ObjAddressKey(addr)
}

func DbKeySCMetaData(addr *address.Address) []byte {
	return ObjAddressKey(addr)
}

func DbKeyVariableState(addr *address.Address) []byte {
	return ObjAddressKey(addr)
}

func DbPrefixState(addr *address.Address, stateIndex uint32) []byte {
	return ObjAddressKey(addr, util.Uint32To4Bytes(stateIndex))
}

func DbKeyStateUpdate(addr *address.Address, stateIndex uint32, batchIndex uint16) []byte {
	return ObjKey(DbPrefixState(addr, stateIndex), util.Uint16To2Bytes(batchIndex))
}

func DbKeyProcessedRequest(reqId *sctransaction.RequestId) []byte {
	return reqId.Bytes()
}
