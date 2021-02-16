package util

import "testing"

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
