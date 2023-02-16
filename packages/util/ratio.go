package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Ratio32 represents a ratio (a:b) between two quantities, expressed as two uint32 values.
type Ratio32 struct {
	A uint32
	B uint32
}

func (r Ratio32) String() string {
	return fmt.Sprintf("%d:%d", r.A, r.B)
}

func Ratio32FromString(s string) (Ratio32, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Ratio32{}, fmt.Errorf("invalid string")
	}
	a, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return Ratio32{}, err
	}
	b, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return Ratio32{}, err
	}
	return Ratio32{A: uint32(a), B: uint32(b)}, nil
}

func (r Ratio32) Bytes() []byte {
	var b [8]byte
	copy(b[:4], Uint32To4Bytes(r.A))
	copy(b[4:], Uint32To4Bytes(r.B))
	return b[:]
}

func Ratio32FromBytes(bytes []byte) (Ratio32, error) {
	if len(bytes) != 8 {
		return Ratio32{}, errors.New("expected bytes length = 8")
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
