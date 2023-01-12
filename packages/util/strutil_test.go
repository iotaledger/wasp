package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCutGently(t *testing.T) {
	t.Log(GentleTruncate("kukukuku", 10))
	t.Log(GentleTruncate("kukukuku", 8))
	t.Log(GentleTruncate("kukukuku", 5))
	t.Log(GentleTruncate("kukukukukuku", 5))
	t.Log(GentleTruncate("kukukukukuku", 6))
	t.Log(GentleTruncate("kukukukukuku", 7))
	t.Log(GentleTruncate("kukukukukuku", 8))
	t.Log(GentleTruncate("kukukukukuku", 9))
	t.Log(GentleTruncate("kukukukukuku", 10))
	t.Log(GentleTruncate("kukukukukuku", 11))
	t.Log(GentleTruncate("kukukukukuku", 12))
	t.Log(GentleTruncate("ku", 1))
	t.Log(GentleTruncate("ku", 5))
	t.Log(GentleTruncate("kuku", 1))
	t.Log(GentleTruncate("kuku", 4))
}

type testShortStringable struct {
	shortString string
}

var _ ShortStringable = &testShortStringable{}

func newTestShortStringable(shortString string) *testShortStringable {
	return &testShortStringable{shortString: shortString}
}

func (tssT *testShortStringable) ShortString() string {
	return tssT.shortString
}

func TestShortStringable(t *testing.T) {
	string1 := "Lorem"
	string2 := "ipsum"
	string3 := "dolor"
	string4 := "sitam"
	slice := []*testShortStringable{
		newTestShortStringable(string1),
		newTestShortStringable(string2),
		newTestShortStringable(string3),
		newTestShortStringable(string4),
	}
	require.Equal(t, SliceShortString(slice[:0]), "[]")
	require.Equal(t, SliceShortString(slice[:1]), "["+string1+"]")
	require.Equal(t, SliceShortString(slice[:2]), "["+string1+"; "+string2+"]")
	require.Equal(t, SliceShortString(slice), "["+string1+"; "+string2+"; "+string3+"; "+string4+"]")
}
