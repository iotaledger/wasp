package iscp

import (
	"math/big"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
)

// KVDecoder is interface with all kind of utility functions extracting and decoding values from the key/value map
type KVDecoder interface {
	kv.KVStoreReader
	GetInt16(key kv.Key, def ...int16) (int16, error)
	MustGetInt16(key kv.Key, def ...int16) int16
	GetUint16(key kv.Key, def ...uint16) (uint16, error)
	MustGetUint16(key kv.Key, def ...uint16) uint16
	GetInt32(key kv.Key, def ...int32) (int32, error)
	MustGetInt32(key kv.Key, def ...int32) int32
	GetUint32(key kv.Key, def ...uint32) (uint32, error)
	MustGetUint32(key kv.Key, def ...uint32) uint32
	GetInt64(key kv.Key, def ...int64) (int64, error)
	MustGetInt64(key kv.Key, def ...int64) int64
	GetUint64(key kv.Key, def ...uint64) (uint64, error)
	MustGetUint64(key kv.Key, def ...uint64) uint64
	GetBool(key kv.Key, def ...bool) (bool, error)
	MustGetBool(key kv.Key, def ...bool) bool
	GetTime(key kv.Key, def ...time.Time) (time.Time, error)
	MustGetTime(key kv.Key, def ...time.Time) time.Time
	GetString(key kv.Key, def ...string) (string, error)
	MustGetString(key kv.Key, def ...string) string
	GetHname(key kv.Key, def ...Hname) (Hname, error)
	MustGetHname(key kv.Key, def ...Hname) Hname
	GetHashValue(key kv.Key, def ...hashing.HashValue) (hashing.HashValue, error)
	MustGetHashValue(key kv.Key, def ...hashing.HashValue) hashing.HashValue
	GetAddress(key kv.Key, def ...iotago.Address) (iotago.Address, error)
	MustGetAddress(key kv.Key, def ...iotago.Address) iotago.Address
	GetRequestID(key kv.Key, def ...RequestID) (RequestID, error)
	MustGetRequestID(key kv.Key, def ...RequestID) RequestID
	GetAgentID(key kv.Key, def ...*AgentID) (*AgentID, error)
	MustGetAgentID(key kv.Key, def ...*AgentID) *AgentID
	GetChainID(key kv.Key, def ...*ChainID) (*ChainID, error)
	MustGetChainID(key kv.Key, def ...*ChainID) *ChainID
	GetBytes(key kv.Key, def ...[]byte) ([]byte, error)
	MustGetBytes(key kv.Key, def ...[]byte) []byte
	GetTokenScheme(key kv.Key, def ...iotago.TokenScheme) (iotago.TokenScheme, error)
	MustGetTokenScheme(key kv.Key, def ...iotago.TokenScheme) iotago.TokenScheme
	GetBigInt(key kv.Key, def ...*big.Int) (*big.Int, error)
	MustGetBigInt(key kv.Key, def ...*big.Int) *big.Int
}
