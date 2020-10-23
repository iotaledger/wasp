package coretypes

import (
	"encoding/binary"
	"strconv"
)

type Uint16 uint16

func (i Uint16) Bytes() []byte {
	ret := make([]byte, 2)
	binary.LittleEndian.PutUint16(ret, (uint16)(i))
	return ret
}

func (i Uint16) String() string {
	return strconv.Itoa(int(i))
}

func NewUint16From2Bytes(data []byte) (Uint16, error) {
	if len(data) != 2 {
		return 0, ErrWrongDataConversion
	}
	return (Uint16)(binary.LittleEndian.Uint16(data)), nil
}
