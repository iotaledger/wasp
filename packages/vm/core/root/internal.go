package root

import (
	"sort"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

func GetContractRegistry(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, StateVarContractRegistry)
}

func GetContractRegistryR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, StateVarContractRegistry)
}

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, nil is returned
func FindContract(state kv.KVStoreReader, hname iscp.Hname) *ContractRecord {
	contractRegistry := GetContractRegistryR(state)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	if retBin != nil {
		ret, err := ContractRecordFromBytes(retBin)
		if err != nil {
			panic(xerrors.Errorf("FindContract: %w", err))
		}
		return ret
	}
	if hname == Contract.Hname() {
		// it happens during bootstrap
		return ContractRecordFromContractInfo(Contract)
	}
	return nil
}

func ContractExists(state kv.KVStoreReader, hname iscp.Hname) bool {
	return GetContractRegistryR(state).MustHasAt(hname.Bytes())
}

// DecodeContractRegistry encodes the whole contract registry from the map into a Go map.
func DecodeContractRegistry(contractRegistry *collections.ImmutableMap) (map[iscp.Hname]*ContractRecord, error) {
	ret := make(map[iscp.Hname]*ContractRecord)
	var err error
	contractRegistry.MustIterate(func(k []byte, v []byte) bool {
		var deploymentHash iscp.Hname
		deploymentHash, err = iscp.HnameFromBytes(k)
		if err != nil {
			return false
		}

		cr, err := ContractRecordFromBytes(v)
		if err != nil {
			return false
		}

		ret[deploymentHash] = cr
		return true
	})
	return ret, err
}

func getBlockContextSubscriptions(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, StateVarBlockContextSubscriptions)
}

func getBlockContextSubscriptionsR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, StateVarBlockContextSubscriptions)
}

func encodeOpenClosePair(openFunc, closeFunc iscp.Hname) []byte {
	return append(codec.EncodeHname(openFunc), codec.EncodeHname(closeFunc)...)
}

func mustDecodeOpenCLosePair(b []byte) (openFunc, closeFunc iscp.Hname) {
	if len(b) != 8 {
		panic("invalid length")
	}
	openFunc = codec.MustDecodeHname(b[0:4])
	closeFunc = codec.MustDecodeHname(b[4:8])
	return
}

func SubscribeBlockContext(state kv.KVStore, contract, openFunc, closeFunc iscp.Hname) {
	getBlockContextSubscriptions(state).MustSetAt(codec.EncodeHname(contract), encodeOpenClosePair(openFunc, closeFunc))
}

type BlockContextSubscription struct {
	Contract  iscp.Hname
	OpenFunc  iscp.Hname
	CloseFunc iscp.Hname
}

// GetBlockContextSubscriptions returns all contracts that are subscribed to block context,
// in deterministic order
func GetBlockContextSubscriptions(state kv.KVStoreReader) []BlockContextSubscription {
	subsMap := getBlockContextSubscriptionsR(state)
	r := make([]BlockContextSubscription, 0, subsMap.MustLen())
	subsMap.MustIterate(func(k []byte, v []byte) bool {
		openFunc, closeFunc := mustDecodeOpenCLosePair(v)
		r = append(r, BlockContextSubscription{
			Contract:  codec.MustDecodeHname(k),
			OpenFunc:  openFunc,
			CloseFunc: closeFunc,
		})
		return true
	})
	sort.Slice(r, func(i, j int) bool { return r[i].Contract < r[j].Contract })
	return r
}
