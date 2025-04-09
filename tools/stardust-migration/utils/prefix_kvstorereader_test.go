package utils

import (
	"sort"
	"testing"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_dict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func TestPrefixKVStoreReader(t *testing.T) {
	s := NewPrefixKVStore(old_dict.New(), nil)

	require.Panics(t, func() {
		s.Iterate("", func(key old_kv.Key, value []byte) bool { return true })
	})
	require.Panics(t, func() {
		s.Iterate("qwe", func(key old_kv.Key, value []byte) bool { return true })
	})

	s.RegisterPrefix("qwe")
	s.RegisterPrefix("qw")

	c := 0
	s.Iterate("qwe", func(key old_kv.Key, value []byte) bool { c++; return true })
	require.Equal(t, 0, c)

	require.Panics(t, func() {
		s.Iterate("asd", func(key old_kv.Key, value []byte) bool { return true })
	})

	require.Nil(t, s.Get(old_kv.Key("qwe")))
	require.Nil(t, s.Get(old_kv.Key("qwe1")))
	require.Nil(t, s.Get(old_kv.Key("asd")))
	require.Nil(t, s.Get(old_kv.Key("asd1")))
	require.Nil(t, s.Get(old_kv.Key("qsd")))
	require.Nil(t, s.Get(old_kv.Key("qsd1")))

	s.Set(old_kv.Key("qwe1"), []byte{1})
	s.Set(old_kv.Key("asd1"), []byte{2})
	s.Set(old_kv.Key("qsd1"), []byte{3})

	require.Equal(t, []byte{1}, s.Get(old_kv.Key("qwe1")))
	require.Equal(t, []byte{2}, s.Get(old_kv.Key("asd1")))
	require.Equal(t, []byte{3}, s.Get(old_kv.Key("qsd1")))

	c = 0
	s.Iterate("qwe", func(key old_kv.Key, value []byte) bool {
		require.Equal(t, old_kv.Key("qwe1"), key)
		require.Equal(t, []byte{1}, value)
		c++
		return true
	})
	require.Equal(t, 1, c)

	require.Panics(t, func() {
		s.Iterate("asd", func(key old_kv.Key, value []byte) bool { return true })
	})
	require.Panics(t, func() {
		s.Iterate("qsd", func(key old_kv.Key, value []byte) bool { return true })
	})
	require.Panics(t, func() {
		s.Iterate("q", func(key old_kv.Key, value []byte) bool { return true })
	})

	s.Set(old_kv.Key("qwe2"), []byte{4})
	s.Set(old_kv.Key("qwe"), []byte{5})
	s.Set(old_kv.Key("qw1"), []byte{6})
	s.Set(old_kv.Key("qw"), []byte{7})

	var keys []string
	var vals [][]byte
	s.Iterate("qwe", func(key old_kv.Key, value []byte) bool {
		keys = append(keys, string(key))
		vals = append(vals, value)
		c++
		return true
	})
	sort.Strings(keys)
	sort.Slice(vals, func(i, j int) bool { return string(vals[i]) < string(vals[j]) })
	require.Equal(t, []string{"qwe", "qwe1", "qwe2"}, keys)
	require.Equal(t, [][]byte{{1}, {4}, {5}}, vals)

	keys = []string{}
	vals = [][]byte{}
	s.Iterate("qw", func(key old_kv.Key, value []byte) bool {
		keys = append(keys, string(key))
		vals = append(vals, value)
		return true
	})
	sort.Strings(keys)
	sort.Slice(vals, func(i, j int) bool { return string(vals[i]) < string(vals[j]) })
	require.Equal(t, []string{"qw", "qw1", "qwe", "qwe1", "qwe2"}, keys)
	require.Equal(t, [][]byte{{1}, {4}, {5}, {6}, {7}}, vals)

	s.Del(old_kv.Key("qw1"))
	require.Nil(t, s.Get(old_kv.Key("qw1")))

	keys = []string{}
	vals = [][]byte{}
	s.Iterate("qw", func(key old_kv.Key, value []byte) bool {
		keys = append(keys, string(key))
		vals = append(vals, value)
		return true
	})
	sort.Strings(keys)
	sort.Slice(vals, func(i, j int) bool { return string(vals[i]) < string(vals[j]) })
	require.Equal(t, []string{"qw", "qwe", "qwe1", "qwe2"}, keys)
	require.Equal(t, [][]byte{{1}, {4}, {5}, {7}}, vals)

	s.Del(old_kv.Key("qwe2"))
	require.Nil(t, s.Get(old_kv.Key("qwe2")))

	keys = []string{}
	vals = [][]byte{}
	s.Iterate("qw", func(key old_kv.Key, value []byte) bool {
		keys = append(keys, string(key))
		vals = append(vals, value)
		return true
	})
	sort.Strings(keys)
	sort.Slice(vals, func(i, j int) bool { return string(vals[i]) < string(vals[j]) })
	require.Equal(t, []string{"qw", "qwe", "qwe1"}, keys)
	require.Equal(t, [][]byte{{1}, {5}, {7}}, vals)

	keys = []string{}
	vals = [][]byte{}
	s.Iterate("qwe", func(key old_kv.Key, value []byte) bool {
		keys = append(keys, string(key))
		vals = append(vals, value)
		return true
	})
	sort.Strings(keys)
	sort.Slice(vals, func(i, j int) bool { return string(vals[i]) < string(vals[j]) })
	require.Equal(t, []string{"qwe", "qwe1"}, keys)
	require.Equal(t, [][]byte{{1}, {5}}, vals)

	s.Set(old_kv.Key("qwe2"), []byte{8})
	require.Equal(t, []byte{8}, s.Get(old_kv.Key("qwe2")))

	keys = []string{}
	vals = [][]byte{}
	s.Iterate("qwe", func(key old_kv.Key, value []byte) bool {
		keys = append(keys, string(key))
		vals = append(vals, value)
		return true
	})
	sort.Strings(keys)
	sort.Slice(vals, func(i, j int) bool { return string(vals[i]) < string(vals[j]) })
	require.Equal(t, []string{"qwe", "qwe1", "qwe2"}, keys)
	require.Equal(t, [][]byte{{1}, {5}, {8}}, vals)
}
