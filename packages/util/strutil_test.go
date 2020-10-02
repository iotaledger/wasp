package util

import "testing"

func TestCutGently(t *testing.T) {
	t.Log(GentleCut("kukukuku", 10))
	t.Log(GentleCut("kukukuku", 8))
	t.Log(GentleCut("kukukuku", 5))
	t.Log(GentleCut("kukukukukuku", 5))
	t.Log(GentleCut("kukukukukuku", 6))
	t.Log(GentleCut("kukukukukuku", 7))
	t.Log(GentleCut("kukukukukuku", 8))
	t.Log(GentleCut("kukukukukuku", 9))
	t.Log(GentleCut("kukukukukuku", 10))
	t.Log(GentleCut("kukukukukuku", 11))
	t.Log(GentleCut("kukukukukuku", 12))
	t.Log(GentleCut("ku", 1))
	t.Log(GentleCut("ku", 5))
	t.Log(GentleCut("kuku", 1))
	t.Log(GentleCut("kuku", 4))
}
