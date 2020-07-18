package kv

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

// Codec is an interface that offers easy conversions between []byte (scalar)
// and other types (including complex) when manipulating a KVStore
type Codec interface {
	RCodec
	WCodec
	GetArray(Key) (*Array, error)
	GetDictionary(Key) (*Dictionary, error)
	// TODO GetTimedLog
}

// MustCodec is like a Codec that automatically panics on error
type MustCodec interface {
	MustRCodec
	WCodec
	GetArray(Key) *MustArray
	GetDictionary(Key) *MustDictionary
	// TODO GetTimedLog
}

// RCodec is an interface that offers easy conversions between []byte and other types when
// manipulating a read-only KVStore
type RCodec interface {
	Has(key Key) (bool, error)
	Get(key Key) ([]byte, error)
	GetString(key Key) (string, bool, error)
	GetInt64(key Key) (int64, bool, error)
	GetAddress(key Key) (*address.Address, bool, error)
	GetHashValue(key Key) (*hashing.HashValue, bool, error)
}

// MustrCodec is like a RCodec that automatically panics on error
type MustRCodec interface {
	Has(key Key) bool
	Get(key Key) []byte
	GetString(key Key) (string, bool)
	GetInt64(key Key) (int64, bool)
	GetAddress(key Key) (*address.Address, bool)
	GetHashValue(key Key) (*hashing.HashValue, bool)
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

type mustcodec struct {
	codec
}

func NewCodec(kv KVStore) Codec {
	return codec{kv: kv}
}

func NewMustCodec(kv KVStore) MustCodec {
	return mustcodec{codec{kv: kv}}
}

func (c codec) GetArray(key Key) (*Array, error) {
	return newArray(c, string(key))
}

func (c mustcodec) GetArray(key Key) *MustArray {
	array, err := c.codec.GetArray(key)
	if err != nil {
		panic(err)
	}
	return newMustArray(array)
}

func (c codec) GetDictionary(key Key) (*Dictionary, error) {
	return newDictionary(c, string(key))
}

func (c mustcodec) GetDictionary(key Key) *MustDictionary {
	d, err := c.codec.GetDictionary(key)
	if err != nil {
		panic(err)
	}
	return newMustDictionary(d)
}

func (c codec) Has(key Key) (bool, error) {
	return c.kv.Has(key)
}

func (c mustcodec) Has(key Key) bool {
	ret, err := c.codec.Has(key)
	if err != nil {
		panic(err)
	}
	return ret
}

func (c codec) Get(key Key) ([]byte, error) {
	return c.kv.Get(key)
}

func (c mustcodec) Get(key Key) []byte {
	ret, err := c.codec.Get(key)
	if err != nil {
		panic(err)
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

func (c mustcodec) GetString(key Key) (string, bool) {
	ret, ok, err := c.codec.GetString(key)
	if err != nil {
		panic(err)
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

func (c mustcodec) GetInt64(key Key) (int64, bool) {
	ret, ok, err := c.codec.GetInt64(key)
	if err != nil {
		panic(err)
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

func (c mustcodec) GetAddress(key Key) (*address.Address, bool) {
	ret, ok, err := c.codec.GetAddress(key)
	if err != nil {
		panic(err)
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

func (c mustcodec) GetHashValue(key Key) (*hashing.HashValue, bool) {
	ret, ok, err := c.codec.GetHashValue(key)
	if err != nil {
		panic(err)
	}
	return ret, ok
}

func (c codec) SetHashValue(key Key, h *hashing.HashValue) {
	c.kv.Set(key, h[:])
}
