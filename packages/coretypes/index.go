package coretypes

import (
	"encoding/binary"
	"strconv"
)

type Int16 uint16

func (i Int16) Bytes() []byte {
	ret := make([]byte, 2)
	binary.LittleEndian.PutUint16(ret, (uint16)(i))
	return ret
}

func (i Int16) String() string {
	return strconv.Itoa(int(i))
}

func NewInt16From2Bytes(data []byte) (Int16, error) {
	if len(data) != 2 {
		return 0, ErrWrongDataConversion
	}
	return (Int16)(binary.LittleEndian.Uint16(data)), nil
}
