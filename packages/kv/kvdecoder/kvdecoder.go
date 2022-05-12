package kvdecoder

import (
	"fmt"
	"math/big"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"golang.org/x/xerrors"
)

type kvdecoder struct {
	kv.KVStoreReader
	log iscp.LogInterface
}

func New(kvReader kv.KVStoreReader, log ...iscp.LogInterface) *kvdecoder {
	var l iscp.LogInterface
	if len(log) > 0 {
		l = log[0]
	}
	return &kvdecoder{kvReader, l}
}

func (p *kvdecoder) check(err error) {
	if err == nil {
		return
	}
	if p.log == nil {
		panic(err)
	}
	p.log.Panicf("%v", err)
}

func (p *kvdecoder) wrapError(key kv.Key, err error) error {
	if err == nil {
		return nil
	}
	return xerrors.Errorf("cannot decode key '%s': %w", key, err)
}

func (p *kvdecoder) GetInt16(key kv.Key, def ...int16) (int16, error) {
	v, err := codec.DecodeInt16(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetInt16(key kv.Key, def ...int16) int16 {
	ret, err := p.GetInt16(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetUint16(key kv.Key, def ...uint16) (uint16, error) {
	v, err := codec.DecodeUint16(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetUint16(key kv.Key, def ...uint16) uint16 {
	ret, err := p.GetUint16(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetInt32(key kv.Key, def ...int32) (int32, error) {
	v, err := codec.DecodeInt32(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetInt32(key kv.Key, def ...int32) int32 {
	ret, err := p.GetInt32(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetUint32(key kv.Key, def ...uint32) (uint32, error) {
	v, err := codec.DecodeUint32(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetUint32(key kv.Key, def ...uint32) uint32 {
	ret, err := p.GetUint32(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetInt64(key kv.Key, def ...int64) (int64, error) {
	v, err := codec.DecodeInt64(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetInt64(key kv.Key, def ...int64) int64 {
	ret, err := p.GetInt64(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetUint64(key kv.Key, def ...uint64) (uint64, error) {
	v, err := codec.DecodeUint64(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetUint64(key kv.Key, def ...uint64) uint64 {
	ret, err := p.GetUint64(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetBool(key kv.Key, def ...bool) (bool, error) {
	v, err := codec.DecodeBool(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetBool(key kv.Key, def ...bool) bool {
	ret, err := p.GetBool(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetTime(key kv.Key, def ...time.Time) (time.Time, error) {
	v, err := codec.DecodeTime(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetTime(key kv.Key, def ...time.Time) time.Time {
	ret, err := p.GetTime(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetString(key kv.Key, def ...string) (string, error) {
	v, err := codec.DecodeString(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetString(key kv.Key, def ...string) string {
	ret, err := p.GetString(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetHname(key kv.Key, def ...iscp.Hname) (iscp.Hname, error) {
	v, err := codec.DecodeHname(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetHname(key kv.Key, def ...iscp.Hname) iscp.Hname {
	ret, err := p.GetHname(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetHashValue(key kv.Key, def ...hashing.HashValue) (hashing.HashValue, error) {
	v, err := codec.DecodeHashValue(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetHashValue(key kv.Key, def ...hashing.HashValue) hashing.HashValue {
	ret, err := p.GetHashValue(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetAddress(key kv.Key, def ...iotago.Address) (iotago.Address, error) {
	v, err := codec.DecodeAddress(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetAddress(key kv.Key, def ...iotago.Address) iotago.Address {
	ret, err := p.GetAddress(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetRequestID(key kv.Key, def ...iscp.RequestID) (iscp.RequestID, error) {
	v, err := codec.DecodeRequestID(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetRequestID(key kv.Key, def ...iscp.RequestID) iscp.RequestID {
	ret, err := p.GetRequestID(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetAgentID(key kv.Key, def ...iscp.AgentID) (iscp.AgentID, error) {
	v, err := codec.DecodeAgentID(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetAgentID(key kv.Key, def ...iscp.AgentID) iscp.AgentID {
	ret, err := p.GetAgentID(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetChainID(key kv.Key, def ...*iscp.ChainID) (*iscp.ChainID, error) {
	v, err := codec.DecodeChainID(p.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *kvdecoder) MustGetChainID(key kv.Key, def ...*iscp.ChainID) *iscp.ChainID {
	ret, err := p.GetChainID(key, def...)
	p.check(err)
	return ret
}

// nil means does not exist
func (p *kvdecoder) GetBytes(key kv.Key, def ...[]byte) ([]byte, error) {
	v := p.MustGet(key)
	if v != nil {
		return v, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("GetBytes: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *kvdecoder) MustGetBytes(key kv.Key, def ...[]byte) []byte {
	ret, err := p.GetBytes(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetTokenScheme(key kv.Key, def ...iotago.TokenScheme) (iotago.TokenScheme, error) {
	v := p.MustGet(key)
	if len(v) > 1 {
		ts, err := iotago.TokenSchemeSelector(uint32(v[0]))
		if err != nil {
			return nil, err
		}
		_, err = ts.Deserialize(v, serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			return nil, err
		}
		return ts, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("GetTokenScheme: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *kvdecoder) MustGetTokenScheme(key kv.Key, def ...iotago.TokenScheme) iotago.TokenScheme {
	ret, err := p.GetTokenScheme(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetTokenTag(key kv.Key, def ...iotago.TokenTag) (iotago.TokenTag, error) {
	v := p.MustGet(key)
	ret, err := codec.DecodeTokenTag(v, def...)
	if err != nil {
		return iotago.TokenTag{}, err
	}
	return ret, nil
}

func (p *kvdecoder) MustGetTokenTag(key kv.Key, def ...iotago.TokenTag) iotago.TokenTag {
	ret, err := p.GetTokenTag(key, def...)
	p.check(err)
	return ret
}

func (p *kvdecoder) GetBigInt(key kv.Key, def ...*big.Int) (*big.Int, error) {
	v := p.MustGet(key)
	if v == nil {
		if len(def) != 0 {
			return def[0], nil
		}
		return nil, fmt.Errorf("GetTokenTag: mandatory parameter '%s' does not exist", key)
	}
	return codec.DecodeBigIntAbs(v)
}

func (p *kvdecoder) MustGetBigInt(key kv.Key, def ...*big.Int) *big.Int {
	ret, err := p.GetBigInt(key, def...)
	p.check(err)
	return ret
}
