package hashing

import (
	"bytes"
	"sort"
)

type sortedHashes []*HashValue

func (s sortedHashes) Len() int {
	return len(s)
}

func (s sortedHashes) Less(i, j int) bool {
	return bytes.Compare(s[i].Bytes(), s[i].Bytes()) < 0
}

func (s sortedHashes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func SortHashes(hashes []*HashValue) {
	sort.Sort(sortedHashes(hashes))
}
