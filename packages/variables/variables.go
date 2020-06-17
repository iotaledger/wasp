package variables

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sort"
)

type Variables interface {
	// TODO tbd
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	GetInt(key string) (int, bool)
	Apply(Variables)
	ForEach(func(key string, value interface{}) bool)
	IsEmpty() bool
	Read(io.Reader) error
	Write(io.Writer) error
	String() string
}

type variables struct {
	m map[string]interface{}
}

// create/clone
func New(vars Variables) Variables {
	return newVars(vars)
}

func newVars(vars Variables) *variables {
	ret := &variables{m: make(map[string]interface{})}
	if vars == nil {
		return ret
	}
	vars.ForEach(func(key string, value interface{}) bool {
		if value != nil {
			ret.m[key] = value
		}
		return true
	})
	return ret
}

func (vr *variables) String() string {
	keys := make([]string, 0)
	for k := range vr.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ret := ""
	for _, key := range keys {
		value := vr.m[key]
		if value == nil {
			ret += fmt.Sprintf("           %s: nil\n", key)
			continue
		}
		switch v := value.(type) {
		case uint16:
			ret += fmt.Sprintf("           %s: %d\n", key, v)
		case uint32:
			ret += fmt.Sprintf("           %s: %d\n", key, v)
		case string:
			ret += fmt.Sprintf("           %s: %s\n", key, v)
		default:
			panic("wrong value type")
		}
	}
	return ret
}

// applying update to variables, var per var.
// newVars value nil mean deleting the variable
func (vr *variables) Apply(upd Variables) {
	upd.ForEach(func(key string, value interface{}) bool {
		if value == nil {
			delete(vr.m, key)
		} else {
			vr.Set(key, value)
		}
		return true
	})
}

func (vr *variables) ForEach(fun func(key string, value interface{}) bool) {
	for k, v := range vr.m {
		if !fun(k, v) {
			return // abort when callback returns false
		}
	}
}

func (vr *variables) IsEmpty() bool {
	return len(vr.m) == 0
}

func (vr *variables) Set(key string, value interface{}) {
	if value == nil {
		vr.m[key] = nil
		return
	}
	switch value.(type) {
	case uint16:
	case uint32:
	case string:
	default:
		panic("wrong value type")
	}
	vr.m[key] = value
}

func (vr *variables) Get(key string) (interface{}, bool) {
	ret, ok := vr.m[key]
	return ret, ok
}

func (vr *variables) GetInt(key string) (int, bool) {
	v, ok := vr.Get(key)
	if !ok {
		return 0, false
	}
	switch vt := v.(type) {
	case uint16:
		return int(vt), true

	case uint32:
		return int(vt), true

	default:
		return 0, false
	}
}

const (
	byteUint16 = iota
	byteUint32
	byteString
	nilValue
)

func (vr *variables) Write(w io.Writer) error {
	ordered := make([]string, 0, len(vr.m))
	for k := range vr.m {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	if err := util.WriteUint16(w, uint16(len(ordered))); err != nil {
		return err
	}
	for _, k := range ordered {
		if err := util.WriteString16(w, k); err != nil {
			return err
		}
		if vr.m[k] == nil {
			if err := util.WriteByte(w, nilValue); err != nil {
				return err
			}
		} else {
			switch tv := vr.m[k].(type) {
			case uint16:
				if err := util.WriteByte(w, byteUint16); err != nil {
					return err
				}
				if err := util.WriteUint16(w, tv); err != nil {
					return err
				}

			case uint32:
				if err := util.WriteByte(w, byteUint32); err != nil {
					return err
				}
				if err := util.WriteUint32(w, tv); err != nil {
					return err
				}

			case string:
				if err := util.WriteByte(w, byteString); err != nil {
					return err
				}
				if err := util.WriteString16(w, tv); err != nil {
					return err
				}

			default:
				panic("wrong type")
			}

		}
	}
	return nil
}

func (vr *variables) Read(r io.Reader) error {
	var num uint16
	err := util.ReadUint16(r, &num)
	if err != nil {
		return err
	}
	var b byte
	var k string
	for i := uint16(0); i < num; i++ {
		if k, err = util.ReadString16(r); err != nil {
			return err
		}
		if b, err = util.ReadByte(r); err != nil {
			return err
		}
		switch b {
		case nilValue:
			vr.Set(k, nil)

		case byteUint16:
			var v uint16
			if err = util.ReadUint16(r, &v); err != nil {
				return err
			}
			vr.Set(k, v)

		case byteUint32:
			var v uint32
			if err = util.ReadUint32(r, &v); err != nil {
				return err
			}
			vr.Set(k, v)

		case byteString:
			var s string
			if s, err = util.ReadString16(r); err != nil {
				return err
			}
			vr.Set(k, s)

		default:
			panic("wrong type byte")
		}
	}
	return nil
}
