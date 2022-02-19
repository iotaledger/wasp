package kv

import (
	"bytes"
	"fmt"
	"sort"
)

type Diff []*DiffElement

type DiffElement struct {
	Key    Key
	Value1 []byte
	Value2 []byte
}

func GetDiff(kvr1, kvr2 KVStoreReader) Diff {
	m := make(map[Key]*DiffElement)
	kvr1.MustIterate("", func(k Key, v1 []byte) bool {
		v2 := kvr2.MustGet(k)
		if !bytes.Equal(v1, v2) {
			m[k] = &DiffElement{
				Key:    k,
				Value1: v1,
				Value2: v2,
			}
		}
		return true
	})
	kvr2.MustIterate("", func(k Key, v2 []byte) bool {
		v1 := kvr1.MustGet(k)
		if !bytes.Equal(v1, v2) {
			m[k] = &DiffElement{
				Key:    k,
				Value1: v1,
				Value2: v2,
			}
		}
		return true
	})
	ret := make(Diff, 0)
	for _, v := range m {
		ret = append(ret, v)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})
	return ret
}

func (d Diff) Dump() string {
	ret := ""
	for _, df := range d {
		ret += fmt.Sprintf("[%3d]%-40s :\n      1: %s\n      2: %s\n", len(df.Key), df.Key, df.Value1, df.Value2)
	}
	return ret
}
