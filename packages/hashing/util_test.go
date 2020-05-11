package hashing

import (
	"testing"
)

func TestLen(t *testing.T) {
	var hashs sortedHashes
	hashs = []*HashValue{
		HashStrings("apple"),
		HashStrings("ball"),
		HashStrings("cat"),
		HashStrings("dog"),
	}
	var length = hashs.Len()
	if length != 4 {
		t.Fatalf("failed to get length")
	}
}

func TestSwap(t *testing.T) {
	var hashs sortedHashes
	hashs = []*HashValue{
		HashStrings("apple"),
		HashStrings("ball"),
		HashStrings("cat"),
		HashStrings("dog"),
	}
	hashs.Swap(0, 3)
	if !hashs[0].Equal(HashStrings("dog")) || !hashs[3].Equal(HashStrings("apple")) {
		t.Fatalf("failed to swap array elements")
	}
}
