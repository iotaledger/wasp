package variables

import (
	"encoding/hex"
	"fmt"
	"github.com/mr-tron/base58"
	"io"
	"sort"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

// Variables represents a key-value map where both keys and values are
// arbitrary byte slices.
//
// Since map cannot have []byte as key, to avoid unnecessary conversions
// between string and []byte, we use string as key data type, but it may
// not necessarily a valid UTF-8 string.
type Variables interface {
	Set(key string, value []byte)
	Del(key string)
	Get(key string) ([]byte, bool)
	IsEmpty() bool
	SetAll(vars Variables)

	ToMap() map[string][]byte

	ForEach(func(key string, value []byte) bool)
	ForEachDeterministic(func(key string, value []byte) bool)

	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	Mutations() MutationSequence

	// TODO: move to a separate interface
	// proposal: to wrap to another interface in a way, that this interface has a call which
	// returns wrapper interface
	SetString(key string, value string)
	GetString(key string) (string, bool)

	SetInt64(key string, value int64)
	GetInt64(key string) (int64, bool, error)
	MustGetInt64(key string) (int64, bool)

	SetAddress(key string, value address.Address)
	GetAddress(key string) (address.Address, bool, error)
	MustGetAddress(key string) (address.Address, bool)

	SetHashValue(key string, value hashing.HashValue)
	GetHashValue(key string) (hashing.HashValue, bool, error)
	MustGetHashValue(key string) (hashing.HashValue, bool)
}

type variables map[string][]byte

// create/clone
func New(vars Variables) Variables {
	ret := make(variables)
	if vars != nil {
		vars.ForEach(func(key string, value []byte) bool {
			ret[key] = value
			return true
		})
	}
	return ret
}

func FromMap(m map[string][]byte) Variables {
	return variables(m)
}

func (vr variables) ToMap() map[string][]byte {
	return variables(vr)
}

func (vr variables) sortedKeys() []string {
	keys := make([]string, 0)
	for k := range vr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (vr variables) String() string {
	ret := ""
	for _, key := range vr.sortedKeys() {
		ret += fmt.Sprintf("           %s: %s (%s)\n", key,
			hex.EncodeToString(vr[key]), base58.Encode(vr[key]))
	}
	return ret
}

func (vr variables) Mutations() MutationSequence {
	ms := NewMutationSequence()
	vr.ForEachDeterministic(func(key string, value []byte) bool {
		ms.Add(NewMutationSet(key, value))
		return true
	})
	return ms
}

// NON DETERMINISTIC!
func (vr variables) ForEach(fun func(key string, value []byte) bool) {
	if vr == nil {
		return
	}
	for k, v := range vr {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (vr variables) ForEachDeterministic(fun func(key string, value []byte) bool) {
	if vr == nil {
		return
	}
	vr.sortedKeys()
	for _, k := range vr.sortedKeys() {
		if !fun(k, vr[k]) {
			return // abort when callback returns false
		}
	}
}

func (vr variables) SetAll(vars Variables) {
	vars.ForEach(func(key string, value []byte) bool {
		vr.Set(key, value)
		return true
	})
}

func (vr variables) IsEmpty() bool {
	return len(vr) == 0
}

func (vr variables) Set(key string, value []byte) {
	if value == nil {
		panic("cannot Set(key, nil), use Del() to remove a key/value")
	}
	vr[key] = value
}

func (vr variables) Del(key string) {
	delete(vr, key)
}

func (vr variables) Get(key string) ([]byte, bool) {
	ret, ok := vr[key]
	return ret, ok
}

func (vr variables) Write(w io.Writer) error {
	keys := vr.sortedKeys()
	if err := util.WriteUint64(w, uint64(len(keys))); err != nil {
		return err
	}
	for _, k := range keys {
		if err := util.WriteString16(w, k); err != nil {
			return err
		}
		if err := util.WriteBytes32(w, vr[k]); err != nil {
			return err
		}
	}
	return nil
}

func (vr variables) Read(r io.Reader) error {
	var num uint64
	err := util.ReadUint64(r, &num)
	if err != nil {
		return err
	}
	for i := uint64(0); i < num; i++ {
		k, err := util.ReadString16(r)
		if err != nil {
			return err
		}
		v, err := util.ReadBytes32(r)
		if err != nil {
			return err
		}
		vr.Set(k, v)
	}
	return nil
}

func (vr variables) SetString(key string, value string) {
	vr.Set(key, []byte(value))
}

func (vr variables) GetString(key string) (string, bool) {
	b, ok := vr.Get(key)
	return string(b), ok
}

func (vr variables) SetInt64(key string, value int64) {
	vr.Set(key, util.Uint64To8Bytes(uint64(value)))
}

func (vr variables) GetInt64(key string) (int64, bool, error) {
	b, ok := vr.Get(key)
	if !ok {
		return 0, false, nil
	}
	if len(b) != 8 {
		return 0, false, fmt.Errorf("variable %s: %v is not an int64", key, b)
	}
	return int64(util.Uint64From8Bytes(b)), true, nil
}

func (vr variables) MustGetInt64(key string) (int64, bool) {
	v, ok, err := vr.GetInt64(key)
	if err != nil {
		panic(err)
	}
	return v, ok
}

func (vr variables) SetAddress(key string, addr address.Address) {
	vr.Set(key, addr[:])
}

func (vr variables) GetAddress(key string) (ret address.Address, ok bool, err error) {
	var b []byte
	b, ok = vr.Get(key)
	if !ok {
		return
	}
	ret, _, err = address.FromBytes(b)
	return
}

func (vr variables) MustGetAddress(key string) (ret address.Address, ok bool) {
	var err error
	ret, ok, err = vr.GetAddress(key)
	if err != nil {
		panic(err)
	}
	return
}

func (vr variables) SetHashValue(key string, h hashing.HashValue) {
	vr.Set(key, h[:])
}

func (vr variables) GetHashValue(key string) (ret hashing.HashValue, ok bool, err error) {
	var b []byte
	b, ok = vr.Get(key)
	if !ok {
		return
	}
	ret, err = hashing.HashValueFromBytes(b)
	return
}

func (vr variables) MustGetHashValue(key string) (ret hashing.HashValue, ok bool) {
	var err error
	ret, ok, err = vr.GetHashValue(key)
	if err != nil {
		panic(err)
	}
	return
}
