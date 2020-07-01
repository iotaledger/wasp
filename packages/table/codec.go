package table

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

// Codec is an interface that offers easy conversions between []byte and other types when
// manipulating a Table
type Codec interface {
	RCodec
	WCodec
}

// RCodec is an interface that offers easy conversions between []byte and other types when
// manipulating a read-only Table
type RCodec interface {
	Get(key Key) ([]byte, error)
	GetString(key Key) (string, bool, error)
	GetInt64(key Key) (int64, bool, error)
	GetAddress(key Key) (*address.Address, bool, error)
	GetHashValue(key Key) (*hashing.HashValue, bool, error)
}

// WCodec is an interface that offers easy conversions between []byte and other types when
// manipulating a write-only Table
type WCodec interface {
	Del(key Key)
	Set(key Key, value []byte)
	SetString(key Key, value string)
	SetInt64(key Key, value int64)
	SetAddress(key Key, value *address.Address)
	SetHashValue(key Key, value *hashing.HashValue)
}

type codec struct {
	kv Table
}

func NewCodec(kv Table) Codec {
	return codec{kv}
}

func NewRCodec(kv Table) RCodec {
	return codec{kv}
}

func (c codec) Get(key Key) ([]byte, error) {
	return c.kv.Get(key)
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

func (c codec) SetHashValue(key Key, h *hashing.HashValue) {
	c.kv.Set(key, h[:])
}
