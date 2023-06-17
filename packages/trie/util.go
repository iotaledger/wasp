package trie

import (
	"fmt"

	"golang.org/x/crypto/blake2b"
)

func concat(par ...[]byte) []byte {
	size := 0
	for _, p := range par {
		size += len(p)
	}
	buf := make([]byte, 0, size)
	for _, p := range par {
		buf = append(buf, p...)
	}
	return buf
}

func blake2b160(data []byte) (ret [HashSizeBytes]byte) {
	hash, _ := blake2b.New(HashSizeBytes, nil)
	if _, err := hash.Write(data); err != nil {
		panic(err)
	}
	copy(ret[:], hash.Sum(nil))
	return
}

func assertf(cond bool, format string, args ...interface{}) {
	if !cond {
		panic(fmt.Sprintf("assertion failed:: "+format, args...))
	}
}

func assertNoError(err error) {
	assertf(err == nil, "error: %v", err)
}
