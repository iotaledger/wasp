package util

import (
	"fmt"

	"golang.org/x/xerrors"
)

// Ratio32 represents a ratio (a:b) between two quantities, expressed as two uint32 values.
type Ratio32 struct {
	A uint32
	B uint32
}

func (r Ratio32) String() string {
	return fmt.Sprintf("%d:%d", r.A, r.B)
}

func (r Ratio32) Bytes() []byte {
	var b [8]byte
	copy(b[:4], Uint32To4Bytes(r.A))
	copy(b[4:], Uint32To4Bytes(r.B))
	return b[:]
}

func Ratio32FromBytes(bytes []byte) (Ratio32, error) {
	if len(bytes) != 8 {
		return Ratio32{}, xerrors.Errorf("expected bytes length = 8")
	}
	a, err := Uint32From4Bytes(bytes[:4])
	if err != nil {
		return Ratio32{}, err
	}
	b, err := Uint32From4Bytes(bytes[4:])
	if err != nil {
		return Ratio32{}, err
	}
	return Ratio32{A: a, B: b}, nil
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
