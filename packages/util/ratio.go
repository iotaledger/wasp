package util

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// A + B
const RatioByteSize = serializer.UInt32ByteSize + serializer.UInt32ByteSize

// Ratio32 represents a ratio (a:b) between two quantities, expressed as two uint32 values.
type Ratio32 struct {
	A uint32 `json:"a" swagger:"min(0),required"`
	B uint32 `json:"b" swagger:"min(0),required"`
}

func Ratio32FromBytes(data []byte) (Ratio32, error) {
	ret, err := bcs.Unmarshal[Ratio32](data)
	if err != nil {
		return ret, err
	}

	if !ret.IsValid() {
		return ret, errors.New("invalid Ratio32")
	}

	return ret, err
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

func (ratio Ratio32) Bytes() []byte {
	return bcs.MustMarshal(&ratio)
}

func (ratio Ratio32) String() string {
	return fmt.Sprintf("%d:%d", ratio.A, ratio.B)
}

func ceil(x, dividend, divisor uint64) uint64 {
	return (x*dividend + (divisor - 1)) / divisor
}

// YFloor64 computes y = floor(x * b / a)
func (ratio Ratio32) YFloor64(x uint64) uint64 {
	return x * uint64(ratio.B) / uint64(ratio.A)
}

// YCeil64 computes y = ceil(x * b / a)
func (ratio Ratio32) YCeil64(x uint64) uint64 {
	return ceil(x, uint64(ratio.B), uint64(ratio.A))
}

// XFloor64 computes x = floor(y * a / b)
func (ratio Ratio32) XFloor64(y uint64) uint64 {
	return y * uint64(ratio.A) / uint64(ratio.B)
}

// XCeil64 computes x = ceil(y * a / b)
func (ratio Ratio32) XCeil64(y uint64) uint64 {
	return ceil(y, uint64(ratio.A), uint64(ratio.B))
}

// TODO: Validate code
// Translated by gpt4o from the based on the functions above.
func ceilBigInt(x, dividend, divisor *big.Int) *big.Int {
	result := new(big.Int)
	temp := new(big.Int).Sub(divisor, big.NewInt(1))
	temp.Mul(x, dividend).Add(temp, temp)
	return result.Div(temp, divisor)
}

// YFloorBigInt computes y = floor(x * b / a)
func (ratio Ratio32) YFloorBigInt(x *big.Int) *big.Int {
	result := new(big.Int)
	b := new(big.Int).SetUint64(uint64(ratio.B))
	a := new(big.Int).SetUint64(uint64(ratio.A))
	result.Mul(x, b).Div(result, a)
	return result
}

// YCeilBigInt computes y = ceil(x * b / a)
func (ratio Ratio32) YCeilBigInt(x *big.Int) *big.Int {
	b := new(big.Int).SetUint64(uint64(ratio.B))
	a := new(big.Int).SetUint64(uint64(ratio.A))
	return ceilBigInt(x, b, a)
}

// XFloorBigInt computes x = floor(y * a / b)
func (ratio Ratio32) XFloorBigInt(y *big.Int) *big.Int {
	result := new(big.Int)
	a := new(big.Int).SetUint64(uint64(ratio.A))
	b := new(big.Int).SetUint64(uint64(ratio.B))
	result.Mul(y, a).Div(result, b)
	return result
}

// XCeilBigInt computes x = ceil(y * a / b)
func (ratio Ratio32) XCeilBigInt(y *big.Int) *big.Int {
	a := new(big.Int).SetUint64(uint64(ratio.A))
	b := new(big.Int).SetUint64(uint64(ratio.B))
	return ceilBigInt(y, a, b)
}

// Set is part of the pflag.Value interface. It accepts a string in the form "a:b".
func (ratio *Ratio32) Set(s string) error {
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
	ratio.A = uint32(a)
	ratio.B = uint32(b)
	return nil
}

// Type is part of the pflag.Value interface.
func (ratio Ratio32) Type() string {
	return "Ratio32"
}

func (ratio Ratio32) HasZeroComponent() bool {
	return ratio.A == 0 || ratio.B == 0
}

func (ratio Ratio32) IsValid() bool {
	return ratio.IsEmpty() || !ratio.HasZeroComponent()
}

func (ratio Ratio32) IsEmpty() bool {
	ZeroGasFee := Ratio32{}
	return ratio == ZeroGasFee
}
