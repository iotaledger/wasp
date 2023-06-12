package util

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// A + B
const RatioByteSize = serializer.UInt32ByteSize + serializer.UInt32ByteSize

// Ratio32 represents a ratio (a:b) between two quantities, expressed as two uint32 values.
type Ratio32 struct {
	A uint32 `json:"a" swagger:"min(0),required"`
	B uint32 `json:"b" swagger:"min(0),required"`
}

func (r Ratio32) String() string {
	return fmt.Sprintf("%d:%d", r.A, r.B)
}

func Ratio32FromString(s string) (ret Ratio32, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return ret, errors.New("invalid Ratio32 string")
	}
	a, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return ret, err
	}
	b, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return ret, err
	}
	ret.A = uint32(a)
	ret.B = uint32(b)
	return ret, nil
}

func (r Ratio32) Bytes() []byte {
	var b [RatioByteSize]byte
	binary.LittleEndian.PutUint32(b[:4], r.A)
	binary.LittleEndian.PutUint32(b[4:], r.B)
	return b[:]
}

func Ratio32FromBytes(bytes []byte) (ret Ratio32, err error) {
	if len(bytes) != RatioByteSize {
		return ret, errors.New("invalid Ratio32 size")
	}
	ret.A = binary.LittleEndian.Uint32(bytes[:4])
	ret.B = binary.LittleEndian.Uint32(bytes[4:])
	return ret, nil
}

func ceil(x, dividend, divisor uint64) uint64 {
	return (x*dividend + (divisor - 1)) / divisor
}

// YFloor64 computes y = floor(x * b / a)
func (r Ratio32) YFloor64(x uint64) uint64 {
	return x * uint64(r.B) / uint64(r.A)
}

// YCeil64 computes y = ceil(x * b / a)
func (r Ratio32) YCeil64(x uint64) uint64 {
	return ceil(x, uint64(r.B), uint64(r.A))
}

// XFloor64 computes x = floor(y * a / b)
func (r Ratio32) XFloor64(y uint64) uint64 {
	return y * uint64(r.A) / uint64(r.B)
}

// XCeil64 computes x = ceil(y * a / b)
func (r Ratio32) XCeil64(y uint64) uint64 {
	return ceil(y, uint64(r.A), uint64(r.B))
}

// Set is part of the pflag.Value interface. It accepts a string in the form "a:b".
func (r *Ratio32) Set(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return errors.New("invalid format for Ratio32")
	}
	a, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return err
	}
	b, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return err
	}
	r.A = uint32(a)
	r.B = uint32(b)
	return nil
}

// Type is part of the pflag.Value interface.
func (r Ratio32) Type() string {
	return "Ratio32"
}

func (r Ratio32) HasZeroComponent() bool {
	return r.A == 0 || r.B == 0
}
