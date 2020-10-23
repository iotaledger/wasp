package coretypes

import (
	"github.com/iotaledger/wasp/packages/hashing"
)

type EntryPointCode Uint32

// NewEntryPointCodeFromFunctionName beware collisions: hash is only 4 bytes!
// must always be checked against the whole table for collisions and adjusted
func NewEntryPointCodeFromFunctionName(funname string) Uint32 {
	ret, _ := NewUint32FromBytes(hashing.HashStrings(funname)[:4])
	return ret
}

func NewEntryPointCodeFromBytes(data []byte) (EntryPointCode, error) {
	r, err := NewUint32FromBytes(data)
	return (EntryPointCode)(r), err
}

func (ep EntryPointCode) Bytes() []byte {
	return (Uint32)(ep).Bytes()
}
