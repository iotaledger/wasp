package kvdecoder

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/colored"
	"golang.org/x/xerrors"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type Decoder struct {
	kv  kv.KVStoreReader
	log iscp.LogInterface
}

func New(kvReader kv.KVStoreReader, log ...iscp.LogInterface) Decoder {
	var l iscp.LogInterface
	if len(log) > 0 {
		l = log[0]
	}
	return Decoder{kvReader, l}
}

func (p *Decoder) check(err error) {
	if err == nil {
		return
	}
	if p.log == nil {
		panic(err)
	}
	p.log.Panicf("%v", err)
}

func (p *Decoder) wrapError(key kv.Key, err error) error {
	if err == nil {
		return nil
	}
	return xerrors.Errorf("cannot decode key '%s': %w", key, err)
}

func (p *Decoder) GetInt16(key kv.Key, def ...int16) (int16, error) {
	v, err := codec.DecodeInt16(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetInt16(key kv.Key, def ...int16) int16 {
	ret, err := p.GetInt16(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetUint16(key kv.Key, def ...uint16) (uint16, error) {
	v, err := codec.DecodeUint16(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetUint16(key kv.Key, def ...uint16) uint16 {
	ret, err := p.GetUint16(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetInt32(key kv.Key, def ...int32) (int32, error) {
	v, err := codec.DecodeInt32(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetInt32(key kv.Key, def ...int32) int32 {
	ret, err := p.GetInt32(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetUint32(key kv.Key, def ...uint32) (uint32, error) {
	v, err := codec.DecodeUint32(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetUint32(key kv.Key, def ...uint32) uint32 {
	ret, err := p.GetUint32(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetInt64(key kv.Key, def ...int64) (int64, error) {
	v, err := codec.DecodeInt64(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetInt64(key kv.Key, def ...int64) int64 {
	ret, err := p.GetInt64(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetUint64(key kv.Key, def ...uint64) (uint64, error) {
	v, err := codec.DecodeUint64(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetUint64(key kv.Key, def ...uint64) uint64 {
	ret, err := p.GetUint64(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetBool(key kv.Key, def ...bool) (bool, error) {
	v, err := codec.DecodeBool(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetBool(key kv.Key, def ...bool) bool {
	ret, err := p.GetBool(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetTime(key kv.Key, def ...time.Time) (time.Time, error) {
	v, err := codec.DecodeTime(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetTime(key kv.Key, def ...time.Time) time.Time {
	ret, err := p.GetTime(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetString(key kv.Key, def ...string) (string, error) {
	v, err := codec.DecodeString(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetString(key kv.Key, def ...string) string {
	ret, err := p.GetString(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetHname(key kv.Key, def ...iscp.Hname) (iscp.Hname, error) {
	v, err := codec.DecodeHname(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetHname(key kv.Key, def ...iscp.Hname) iscp.Hname {
	ret, err := p.GetHname(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetHashValue(key kv.Key, def ...hashing.HashValue) (hashing.HashValue, error) {
	v, err := codec.DecodeHashValue(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetHashValue(key kv.Key, def ...hashing.HashValue) hashing.HashValue {
	ret, err := p.GetHashValue(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetAddress(key kv.Key, def ...ledgerstate.Address) (ledgerstate.Address, error) {
	v, err := codec.DecodeAddress(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetAddress(key kv.Key, def ...ledgerstate.Address) ledgerstate.Address {
	ret, err := p.GetAddress(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetRequestID(key kv.Key, def ...iscp.RequestID) (iscp.RequestID, error) {
	v, err := codec.DecodeRequestID(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetRequestID(key kv.Key, def ...iscp.RequestID) iscp.RequestID {
	ret, err := p.GetRequestID(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetAgentID(key kv.Key, def ...*iscp.AgentID) (*iscp.AgentID, error) {
	v, err := codec.DecodeAgentID(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetAgentID(key kv.Key, def ...*iscp.AgentID) *iscp.AgentID {
	ret, err := p.GetAgentID(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetChainID(key kv.Key, def ...*iscp.ChainID) (*iscp.ChainID, error) {
	v, err := codec.DecodeChainID(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetChainID(key kv.Key, def ...*iscp.ChainID) *iscp.ChainID {
	ret, err := p.GetChainID(key, def...)
	p.check(err)
	return ret
}

func (p *Decoder) GetColor(key kv.Key, def ...colored.Color) (colored.Color, error) {
	v, err := codec.DecodeColor(p.kv.MustGet(key), def...)
	return v, p.wrapError(key, err)
}

func (p *Decoder) MustGetColor(key kv.Key, def ...colored.Color) colored.Color {
	ret, err := p.GetColor(key, def...)
	p.check(err)
	return ret
}

// nil means does not exist
func (p *Decoder) GetBytes(key kv.Key, def ...[]byte) ([]byte, error) {
	v := p.kv.MustGet(key)
	if v != nil {
		return v, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("GetBytes: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetBytes(key kv.Key, def ...[]byte) []byte {
	ret, err := p.GetBytes(key, def...)
	p.check(err)
	return ret
}
