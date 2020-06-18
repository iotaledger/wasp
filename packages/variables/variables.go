package variables

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sort"
)

type Variables interface {
	Set(key string, value []byte)
	Del(key string)
	Get(key string) ([]byte, bool)
	IsEmpty() bool

	ForEach(func(key string, value []byte) bool)

	Read(io.Reader) error
	Write(io.Writer) error

	String() string

	// TODO: move to a separate interface
	SetString(key string, value string)
	GetString(key string) (string, bool)
	SetInt64(key string, value int64)
	GetInt64(key string) (int64, bool)
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
		ret += fmt.Sprintf("           %s: %v\n", key, vr[key])
	}
	return ret
}

func (vr variables) ForEach(fun func(key string, value []byte) bool) {
	for k, v := range vr {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
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

func (vr variables) GetInt64(key string) (int64, bool) {
	b, ok := vr.Get(key)
	if !ok {
		return 0, false
	}
	return int64(util.Uint64From8Bytes(b)), true
}
