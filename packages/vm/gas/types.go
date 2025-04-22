// Package gas implements gas fees for the vm
package gas

import (
	"errors"
	"fmt"
)

type BurnCode uint16

type BurnFunction func(x uint64) uint64

type BurnCodeRecord struct {
	Name string
	BurnFunction
}

type BurnTable map[BurnCode]BurnCodeRecord

var ErrUnknownBurnCode = errors.New("unknown gas burn code")

func (c BurnCode) Name() string {
	r, ok := burnTable[c]
	if !ok {
		return "(undef)"
	}
	return r.Name
}

func BurnCodeFromName(name string) BurnCode {
	for burnCode := range burnTable {
		if burnCode.Name() == name {
			return burnCode
		}
	}
	panic(fmt.Sprintf("name %s not exist", name))
}
