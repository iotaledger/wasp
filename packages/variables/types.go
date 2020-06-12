package variables

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"sort"
)

type Variables interface {
	// TODO tbd
	Set(key interface{}, value interface{})
	Get(key interface{}) (interface{}, bool)
	Apply(Variables)
	ForEach(func(key interface{}, value interface{}) bool) bool
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
	vars.ForEach(func(key interface{}, value interface{}) bool {
		if value != nil {
			ret.m[key.(string)] = value
		}
		return true
	})
	return ret
}

func (vr *variables) String() string {
	ret := ""
	vr.ForEach(func(key interface{}, value interface{}) bool {
		k := key.(string)
		switch v := value.(type) {
		case uint16:
			ret += fmt.Sprintf("%s: %d\n", k, v)
		case uint32:
			ret += fmt.Sprintf("%s: %d\n", k, v)
		case string:
			ret += fmt.Sprintf("%s: %s\n", k, v)
		default:
			panic("wrong value type")
		}
		return true
	})
	return ret
}

// applying update to variables, var per var.
// newVars value nil mean deleting the variable
func (vr *variables) Apply(upd Variables) {
	upd.ForEach(func(key interface{}, value interface{}) bool {
		skey := key.(string)
		if value == nil {
			delete(vr.m, skey)
		} else {
			vr.Set(skey, value)
		}
		return false
	})
}

func (vr *variables) ForEach(fun func(key interface{}, value interface{}) bool) bool {
	for k, v := range vr.m {
		if fun(k, v) {
			return true // aborted
		}
	}
	return false
}

func (vr *variables) Set(key interface{}, value interface{}) {
	if value == nil {
		delete(vr.m, key.(string))
		return
	}
	switch value.(type) {
	case uint16:
	case uint32:
	case string:
	default:
		panic("wrong value type")
	}
	vr.m[key.(string)] = value
}

func (vr *variables) Get(key interface{}) (interface{}, bool) {
	ret, ok := vr.m[key.(string)]
	return ret, ok
}

const (
	byteUint16 = iota
	byteUint32
	byteString
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
