package hashing

import (
	"bytes"
	"io"
	"sort"
)

func Bytes(obj interface{ Write(io.Writer) error }) ([]byte, error) {
	var buf bytes.Buffer
	if err := obj.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MustBytes(obj interface{ Write(io.Writer) error }) []byte {
	ret, err := Bytes(obj)
	if err != nil {
		panic(err)
	}
	return ret
}

func GetHashValue(obj interface{ Write(io.Writer) error }) HashValue {
	return *HashData(MustBytes(obj))
}

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
