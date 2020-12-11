package datatypes

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
)

type Array struct {
	kv   kv.KVStore
	name string
}

type MustArray struct {
	array Array
}

func NewArray(kv kv.KVStore, name string) *Array {
	return &Array{
		kv:   kv,
		name: name,
	}
}

func NewMustArray(kv kv.KVStore, name string) *MustArray {
	return NewArray(kv, name).Must()
}

const (
	arraySizeKeyCode = byte(0)
	arrayElemKeyCode = byte(1)
)

func (a *Array) Must() *MustArray {
	return &MustArray{*a}
}

func (l *Array) getSizeKey() kv.Key {
	return ArraySizeKey(l.name)
}

func ArraySizeKey(name string) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arraySizeKeyCode)
	return kv.Key(buf.Bytes())
}

func (l *Array) getElemKey(idx uint16) kv.Key {
	return ArrayElemKey(l.name, idx)
}

func ArrayElemKey(name string, idx uint16) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(arrayElemKeyCode)
	_ = util.WriteUint16(&buf, idx)
	return kv.Key(buf.Bytes())
}

// ArrayRangeKeys returns the KVStore keys for the items between [from, to) (`to` being not inclusive),
// assuming it has `length` elements.
func ArrayRangeKeys(name string, length uint16, from uint16, to uint16) []kv.Key {
	keys := make([]kv.Key, 0)
	if to >= from {
		for i := from; i < to && i < length; i++ {
			keys = append(keys, ArrayElemKey(name, i))
		}
	}
	return keys
}

func (a *Array) setSize(n uint16) {
	if n == 0 {
		a.kv.Del(a.getSizeKey())
	} else {
		a.kv.Set(a.getSizeKey(), util.Uint16To2Bytes(n))
	}
}

func (a *Array) addToSize(amount int) (uint16, error) {
	prevSize, err := a.Len()
	if err != nil {
		return 0, err
	}
	a.setSize(uint16(int(prevSize) + amount))
	return prevSize, nil
}

// Len == 0/empty/non-existent are equivalent
func (l *Array) Len() (uint16, error) {
	v, err := l.kv.Get(l.getSizeKey())
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	return util.MustUint16From2Bytes(v), nil
}

func (a *MustArray) Len() uint16 {
	n, err := a.array.Len()
	if err != nil {
		panic(err)
	}
	return n
}

// adds to the end of the list
func (l *Array) Push(value []byte) error {
	prevSize, err := l.addToSize(1)
	if err != nil {
		return err
	}
	k := l.getElemKey(prevSize)
	l.kv.Set(k, value)
	return nil
}

func (a *MustArray) Push(value []byte) {
	err := a.array.Push(value)
	if err != nil {
		panic(err)
	}
}

func (l *Array) Extend(other *Array) error {
	otherLen, err := other.Len()
	if err != nil {
		return err
	}
	for i := uint16(0); i < otherLen; i++ {
		v, _ := other.GetAt(i)
		err = l.Push(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *MustArray) Extend(other *MustArray) {
	a.array.Extend(&other.array)
}

// TODO implement with DelPrefix
func (l *Array) Erase() error {
	n, err := l.Len()
	if err != nil {
		return err
	}
	for i := uint16(0); i < n; i++ {
		l.kv.Del(l.getElemKey(i))
	}
	l.setSize(0)
	return nil
}

func (a *MustArray) Erase() {
	a.array.Erase()
}

func (l *Array) GetAt(idx uint16) ([]byte, error) {
	n, err := l.Len()
	if err != nil {
		return nil, err
	}
	if idx >= n {
		return nil, fmt.Errorf("index %d out of range for array of len %d", idx, n)
	}
	ret, err := l.kv.Get(l.getElemKey(idx))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (a *MustArray) GetAt(idx uint16) []byte {
	ret, err := a.array.GetAt(idx)
	if err != nil {
		panic(err)
	}
	return ret
}

func (l *Array) SetAt(idx uint16, value []byte) error {
	n, err := l.Len()
	if err != nil {
		return err
	}
	if idx >= n {
		return fmt.Errorf("index %d out of range for array of len %d", idx, n)
	}
	l.kv.Set(l.getElemKey(idx), value)
	return nil
}

func (a *MustArray) SetAt(idx uint16, value []byte) error {
	return a.array.SetAt(idx, value)
}
