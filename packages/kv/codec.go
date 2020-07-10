package kv

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

// Codec is an interface that offers easy conversions between []byte and other types when
// manipulating a KVStore
type Codec interface {
	RCodec
	MustRCodec
	WCodec
	GetArray(Key) (*Array, error)
	MustGetArray(Key) *Array
	GetDictionary(Key) (Dictionary, error)
	MustGetDictionary(Key) Dictionary
	// TODO GetTimedLog
}

// RCodec is an interface that offers easy conversions between []byte and other types when
// manipulating a read-only KVStore
type RCodec interface {
	Get(key Key) ([]byte, error)
	GetString(key Key) (string, bool, error)
	GetInt64(key Key) (int64, bool, error)
	GetAddress(key Key) (*address.Address, bool, error)
	GetHashValue(key Key) (*hashing.HashValue, bool, error)
}

type MustRCodec interface {
	MustGet(key Key) []byte
	MustGetString(key Key) (string, bool)
	MustGetInt64(key Key) (int64, bool)
	MustGetAddress(key Key) (*address.Address, bool)
	MustGetHashValue(key Key) (*hashing.HashValue, bool)
}

// WCodec is an interface that offers easy conversions between []byte and other types when
// manipulating a write-only KVStore
type WCodec interface {
	Del(key Key)
	Set(key Key, value []byte)
	SetString(key Key, value string)
	SetInt64(key Key, value int64)
	SetAddress(key Key, value *address.Address)
	SetHashValue(key Key, value *hashing.HashValue)
}

type codec struct {
	kv KVStore
}

func NewCodec(kv KVStore) Codec {
	return codec{kv}
}

func NewRCodec(kv KVStore) RCodec {
	return codec{kv}
}

func (c codec) GetArray(key Key) (*Array, error) {
	return NewArray(c, string(key))
}

func (c codec) MustGetArray(key Key) *Array {
	ret, err := NewArray(c, string(key))
	if err != nil {
		panic(err)
	}
	return ret
}

func (c codec) GetDictionary(key Key) (Dictionary, error) {
	return newDict(c, string(key))
}

func (c codec) MustGetDictionary(key Key) Dictionary {
	ret, err := c.GetDictionary(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (c codec) Get(key Key) ([]byte, error) {
	return c.kv.Get(key)
}

func (c codec) MustGet(key Key) []byte {
	ret, err := c.Get(key)
	if err != nil {
		panic(err) // or invalidation of the whole virtual state
	}
	return ret
}

func (c codec) Del(key Key) {
	c.kv.Del(key)
}

func (c codec) Set(key Key, value []byte) {
	c.kv.Set(key, value)
}

func (c codec) GetString(key Key) (string, bool, error) {
	b, err := c.kv.Get(key)
	if err != nil || b == nil {
		return "", false, err
	}
	return string(b), true, nil
}

func (c codec) MustGetString(key Key) (string, bool) {
	ret, ok, err := c.GetString(key)
	if err != nil {
		panic(err) // or invalidation of the whole virtual state
	}
	return ret, ok
}

func (c codec) SetString(key Key, value string) {
	c.kv.Set(key, []byte(value))
}

func (c codec) GetInt64(key Key) (int64, bool, error) {
	b, err := c.kv.Get(key)
	if err != nil || b == nil {
		return 0, false, err
	}
	if len(b) != 8 {
		return 0, false, fmt.Errorf("variable %s: %v is not an int64", key, b)
	}
	return int64(util.Uint64From8Bytes(b)), true, nil
}

func (c codec) MustGetInt64(key Key) (int64, bool) {
	ret, ok, err := c.GetInt64(key)
	if err != nil {
		panic(err) // or invalidation of the whole virtual state
	}
	return ret, ok
}

func (c codec) SetInt64(key Key, value int64) {
	c.kv.Set(key, util.Uint64To8Bytes(uint64(value)))
}

func (c codec) GetAddress(key Key) (*address.Address, bool, error) {
	b, err := c.kv.Get(key)
	if err != nil || b == nil {
		return nil, false, err
	}
	ret, _, err := address.FromBytes(b)
	return &ret, true, nil
}

func (c codec) MustGetAddress(key Key) (*address.Address, bool) {
	ret, ok, err := c.GetAddress(key)
	if err != nil {
		panic(err) // or invalidate of the whole virtual state
	}
	return ret, ok
}

func (c codec) SetAddress(key Key, addr *address.Address) {
	c.kv.Set(key, addr.Bytes())
}

func (c codec) GetHashValue(key Key) (*hashing.HashValue, bool, error) {
	var b []byte
	b, err := c.kv.Get(key)
	if err != nil || b == nil {
		return nil, false, err
	}
	ret, err := hashing.HashValueFromBytes(b)
	return &ret, err == nil, err
}

func (c codec) MustGetHashValue(key Key) (*hashing.HashValue, bool) {
	ret, ok, err := c.GetHashValue(key)
	if err != nil {
		panic(err) // or invalidate of the whole virtual state
	}
	return ret, ok
}

func (c codec) SetHashValue(key Key, h *hashing.HashValue) {
	c.kv.Set(key, h[:])
}
