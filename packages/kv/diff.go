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

func GetAllKeys(pref Key, kvr1, kvr2 KVStoreReader) map[Key]struct{} {
	ret := make(map[Key]struct{}, 0)
	kvr1.MustIterate(pref, func(k Key, _ []byte) bool {
		ret[k] = struct{}{}
		return true
	})
	kvr2.MustIterate(pref, func(k Key, _ []byte) bool {
		ret[k] = struct{}{}
		return true
	})
	return ret
}

func GetDiffKeyValues(kvr1, kvr2 KVStoreReader) map[Key]struct{} {
	ret := make(map[Key]struct{})
	kvr1.MustIterate("", func(k Key, v1 []byte) bool {
		v2 := kvr2.MustGet(k)
		if !bytes.Equal(v1, v2) {
			ret[k] = struct{}{}
		}
		return true
	})
	kvr2.MustIterate("", func(k Key, v2 []byte) bool {
		v1 := kvr1.MustGet(k)
		if !bytes.Equal(v1, v2) {
			ret[k] = struct{}{}
		}
		return true
	})
	return ret
}

func GetDiffKeys(kvr1, kvr2 KVStoreReader) map[Key]struct{} {
	ret := make(map[Key]struct{})
	kvr1.MustIterateKeys("", func(k Key) bool {
		if !kvr2.MustHas(k) {
			ret[k] = struct{}{}
		}
		return true
	})
	kvr2.MustIterateKeys("", func(k Key) bool {
		if !kvr1.MustHas(k) {
			ret[k] = struct{}{}
		}
		return true
	})
	return ret
}

func NumKeys(kvr KVMustIterator) int {
	ret := 0
	kvr.MustIterateKeys("", func(_ Key) bool {
		ret++
		return true
	})
	return ret
}

func GetAllKeysWithDiff(kvr1, kvr2 KVStoreReader) string {
	allkeys := GetAllKeys("", kvr1, kvr2)
	diffKeys := GetDiffKeyValues(kvr1, kvr2)
	keys := make([]Key, 0)
	for k := range allkeys {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	ret := ""
	for _, k := range keys {
		if _, ok := diffKeys[k]; ok {
			ret += fmt.Sprintf("   >>>>>>>>>>> %s\n", string(k))
		} else {
			ret += fmt.Sprintf("               %s\n", string(k))
		}
	}
	return ret
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

func (d Diff) DumpDiff() string {
	ret := ""
	for _, df := range d {
		ret += fmt.Sprintf("[%3d]%-40s :\n      1: %s\n      2: %s\n", len(df.Key), df.Key, df.Value1, df.Value2)
	}
	return ret
}

func DumpKeySet(s map[Key]struct{}) []string {
	ret := make([]string, 0)
	for k := range s {
		ret = append(ret, string(k))
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}
