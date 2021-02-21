package kvdecoder

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type decoder struct {
	kv  kv.KVStoreReader
	log coretypes.LogInterface
}

func New(kv kv.KVStoreReader, log ...coretypes.LogInterface) decoder {
	var l coretypes.LogInterface
	if len(log) > 0 {
		l = log[0]
	}
	return decoder{kv, l}
}

func (p *decoder) panic(err error) {
	if p.log == nil {
		panic(err)
	}
	p.log.Panicf("%v", err)
}

func (p *decoder) GetInt64(key kv.Key, def ...int64) (int64, error) {
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
func (p *decoder) MustGetInt64(key kv.Key, def ...int64) int64 {
	ret, err := p.GetInt64(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetString(key kv.Key, def ...string) (string, error) {
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

func (p *decoder) MustGetString(key kv.Key, def ...string) string {
	ret, err := p.GetString(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetHname(key kv.Key, def ...coretypes.Hname) (coretypes.Hname, error) {
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

func (p *decoder) MustGetHname(key kv.Key, def ...coretypes.Hname) coretypes.Hname {
	ret, err := p.GetHname(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetHashValue(key kv.Key, def ...hashing.HashValue) (hashing.HashValue, error) {
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

func (p *decoder) MustGetHashValue(key kv.Key, def ...hashing.HashValue) hashing.HashValue {
	ret, err := p.GetHashValue(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetAddress(key kv.Key, def ...address.Address) (address.Address, error) {
	v, exists, err := codec.DecodeAddress(p.kv.MustGet(key))
	if err != nil {
		return address.Address{}, fmt.Errorf("GetAddress: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return address.Address{}, fmt.Errorf("GetAddress: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *decoder) MustGetAddress(key kv.Key, def ...address.Address) address.Address {
	ret, err := p.GetAddress(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetContractID(key kv.Key, def ...coretypes.ContractID) (coretypes.ContractID, error) {
	v, exists, err := codec.DecodeContractID(p.kv.MustGet(key))
	if err != nil {
		return coretypes.ContractID{}, fmt.Errorf("GetContractID: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return coretypes.ContractID{}, fmt.Errorf("GetContractID: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *decoder) MustGetContractID(key kv.Key, def ...coretypes.ContractID) coretypes.ContractID {
	ret, err := p.GetContractID(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetAgentID(key kv.Key, def ...coretypes.AgentID) (coretypes.AgentID, error) {
	v, exists, err := codec.DecodeAgentID(p.kv.MustGet(key))
	if err != nil {
		return coretypes.AgentID{}, fmt.Errorf("GetAgentID: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return coretypes.AgentID{}, fmt.Errorf("GetAgentID: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *decoder) MustGetAgentID(key kv.Key, def ...coretypes.AgentID) coretypes.AgentID {
	ret, err := p.GetAgentID(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetChainID(key kv.Key, def ...coretypes.ChainID) (coretypes.ChainID, error) {
	v, exists, err := codec.DecodeChainID(p.kv.MustGet(key))
	if err != nil {
		return coretypes.ChainID{}, fmt.Errorf("GetChainID: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return coretypes.ChainID{}, fmt.Errorf("GetChainID: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *decoder) MustGetChainID(key kv.Key, def ...coretypes.ChainID) coretypes.ChainID {
	ret, err := p.GetChainID(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

func (p *decoder) GetColor(key kv.Key, def ...balance.Color) (balance.Color, error) {
	v, exists, err := codec.DecodeColor(p.kv.MustGet(key))
	if err != nil {
		return balance.Color{}, fmt.Errorf("GetColor: decoding parameter '%s': %v", key, err)
	}
	if exists {
		return v, nil
	}
	if len(def) == 0 {
		return balance.Color{}, fmt.Errorf("GetColor: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *decoder) MustGetColor(key kv.Key, def ...balance.Color) balance.Color {
	ret, err := p.GetColor(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}

// nil means does not exist
func (p *decoder) GetBytes(key kv.Key, def ...[]byte) ([]byte, error) {
	v := p.kv.MustGet(key)
	if v != nil {
		return v, nil
	}
	if len(def) == 0 {
		return nil, fmt.Errorf("MustGetBytes: mandatory parameter '%s' does not exist", key)
	}
	return def[0], nil
}

func (p *decoder) MustGetBytes(key kv.Key, def ...[]byte) []byte {
	ret, err := p.GetBytes(key, def...)
	if err != nil {
		p.panic(err)
	}
	return ret
}
