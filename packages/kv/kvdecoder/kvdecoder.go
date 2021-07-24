package kvdecoder

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/iscp/color"
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

func (p *Decoder) panic(err error) {
	if p.log == nil {
		panic(err)
	}
	p.log.Panicf("%v", err)
}

func (p *Decoder) GetInt16(key kv.Key, def ...int16) (int16, error) {
	v, exists, err := codec.DecodeInt16(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetInt16: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetInt16: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetInt16(key kv.Key, def ...int16) int16 {
	ret, err := p.GetInt16(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetUint16(key kv.Key, def ...uint16) (uint16, error) {
	v, exists, err := codec.DecodeUint16(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetUint16: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetUint16: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetUint16(key kv.Key, def ...uint16) uint16 {
	ret, err := p.GetUint16(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetInt32(key kv.Key, def ...int32) (int32, error) {
	v, exists, err := codec.DecodeInt32(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetInt32: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetInt32: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetInt32(key kv.Key, def ...int32) int32 {
	ret, err := p.GetInt32(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetUint32(key kv.Key, def ...uint32) (uint32, error) {
	v, exists, err := codec.DecodeUint32(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetUint32: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetUint32: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetUint32(key kv.Key, def ...uint32) uint32 {
	ret, err := p.GetUint32(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetInt64(key kv.Key, def ...int64) (int64, error) {
	v, exists, err := codec.DecodeInt64(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetInt64: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetInt64: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetInt64(key kv.Key, def ...int64) int64 {
	ret, err := p.GetInt64(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetUint64(key kv.Key, def ...uint64) (uint64, error) {
	v, exists, err := codec.DecodeUint64(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetUint64: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetUint64: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetUint64(key kv.Key, def ...uint64) uint64 {
	ret, err := p.GetUint64(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetTime(key kv.Key, def ...time.Time) (time.Time, error) {
	v, exists, err := codec.DecodeTime(p.kv.MustGet(key))
	if err != nil {
		return time.Time{}, fmt.Errorf("GetTime: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return time.Time{}, fmt.Errorf("GetUint32: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetTime(key kv.Key, def ...time.Time) time.Time {
	ret, err := p.GetTime(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetString(key kv.Key, def ...string) (string, error) {
	v, exists, err := codec.DecodeString(p.kv.MustGet(key))
	if err != nil {
		return "", fmt.Errorf("GetString: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return "", fmt.Errorf("GetString: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetString(key kv.Key, def ...string) string {
	ret, err := p.GetString(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetHname(key kv.Key, def ...iscp.Hname) (iscp.Hname, error) {
	v, exists, err := codec.DecodeHname(p.kv.MustGet(key))
	if err != nil {
		return 0, fmt.Errorf("GetHname: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return 0, fmt.Errorf("GetHname: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetHname(key kv.Key, def ...iscp.Hname) iscp.Hname {
	ret, err := p.GetHname(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetHashValue(key kv.Key, def ...hashing.HashValue) (hashing.HashValue, error) {
	v, exists, err := codec.DecodeHashValue(p.kv.MustGet(key))
	if err != nil {
		return [32]byte{}, fmt.Errorf("GetHashValue: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return hashing.NilHash, fmt.Errorf("GetHashValue: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetHashValue(key kv.Key, def ...hashing.HashValue) hashing.HashValue {
	ret, err := p.GetHashValue(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetAddress(key kv.Key, def ...ledgerstate.Address) (ledgerstate.Address, error) {
	v, exists, err := codec.DecodeAddress(p.kv.MustGet(key))
	if err != nil {
		return nil, fmt.Errorf("GetAddress: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("GetAddress: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetAddress(key kv.Key, def ...ledgerstate.Address) ledgerstate.Address {
	ret, err := p.GetAddress(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetRequestID(key kv.Key, def ...iscp.RequestID) (iscp.RequestID, error) {
	v, exists, err := codec.DecodeRequestID(p.kv.MustGet(key))
	if err != nil {
		return iscp.RequestID{}, fmt.Errorf("GetRequestID: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return iscp.RequestID{}, fmt.Errorf("GetRequestID: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetRequestID(key kv.Key, def ...iscp.RequestID) iscp.RequestID {
	ret, err := p.GetRequestID(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetAgentID(key kv.Key, def ...iscp.AgentID) (*iscp.AgentID, error) {
	v, exists, err := codec.DecodeAgentID(p.kv.MustGet(key))
	if err != nil {
		return nil, fmt.Errorf("GetAgentID: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return &v, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("GetAgentID: mandatory parameter '%s' does not exist", key)
	}
	r := def[0]
	return &r, nil
}

func (p *Decoder) MustGetAgentID(key kv.Key, def ...iscp.AgentID) *iscp.AgentID {
	ret, err := p.GetAgentID(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetChainID(key kv.Key, def ...iscp.ChainID) (*iscp.ChainID, error) {
	v, exists, err := codec.DecodeChainID(p.kv.MustGet(key))
	if err != nil {
		return nil, fmt.Errorf("GetChainID: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return &v, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("GetChainID: mandatory parameter '%s' does not exist", key)
	}
	r := def[0]
	return &r, nil
}

func (p *Decoder) MustGetChainID(key kv.Key, def ...iscp.ChainID) *iscp.ChainID {
	ret, err := p.GetChainID(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

// Deprecated: use GetColor instead
func (p *Decoder) GetColorLedgerstate(key kv.Key, def ...ledgerstate.Color) (ledgerstate.Color, error) {
	v, exists, err := codec.DecodeColorLedgerstate(p.kv.MustGet(key))
	if err != nil {
		return ledgerstate.Color{}, fmt.Errorf("GetColorLedgerstate: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return ledgerstate.Color{}, fmt.Errorf("GetColorLedgerstate: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

// Deprecated: use MustGetColor instead
func (p *Decoder) MustGetColorLedgerstate(key kv.Key, def ...ledgerstate.Color) ledgerstate.Color {
	ret, err := p.GetColorLedgerstate(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *Decoder) GetColor(key kv.Key, def ...color.Color) (color.Color, error) {
	v, exists, err := codec.DecodeColor(p.kv.MustGet(key))
	if err != nil {
		return color.Color{}, xerrors.Errorf("GetColor: decoding parameter '%s': %w", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return color.Color{}, xerrors.Errorf("GetColor: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *Decoder) MustGetColor(key kv.Key, def ...color.Color) color.Color {
	ret, err := p.GetColor(key, def...)
	if err != nil {
		p.panic(err)
	}
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
	if err != nil {
		p.panic(err)
	}
	return ret
}
