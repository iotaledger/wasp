// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"io"
	"math"

	"fortio.org/safecast"
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type BitVector interface {
	SetBits(positions []int) BitVector
	AsInts() []int
	Bytes() []byte
	Read(r io.Reader) error
	Write(w io.Writer) error
}

func init() {
	bcs.RegisterEnumType1[BitVector, *fixBitVector]()
}

type fixBitVector struct {
	size uint16 `bcs:"export"`
	data []byte `bcs:"export"`
}

func NewFixedSizeBitVector(size uint16) BitVector {
	return &fixBitVector{size: size, data: make([]byte, (size+7)/8)}
}

func FixedSizeBitVectorFromBytes(data []byte) (BitVector, error) {
	return rwutil.ReadFromBytes(data, new(fixBitVector))
}

func (b *fixBitVector) Bytes() []byte {
	return rwutil.WriteToBytes(b)
}

func (b *fixBitVector) SetBits(positions []int) BitVector {
	for _, p := range positions {
		bytePos, bitMask := b.bitMask(p)
		b.data[bytePos] |= bitMask
	}
	return b
}

func (b *fixBitVector) AsInts() []int {
	ints := make([]int, 0, b.size)
	for i := 0; i < int(b.size); i++ {
		bytePos, bitMask := b.bitMask(i)
		if b.data[bytePos]&bitMask != 0 {
			ints = append(ints, i)
		}
	}
	return ints
}

func (b *fixBitVector) bitMask(position int) (int, byte) {
	positionU32, err := safecast.Convert[uint32](position)
	if err != nil {
		panic("position is too large for uint32")
	}
	sizeU32, err := safecast.Convert[uint32](b.size)
	if err != nil {
		panic("size is too large for uint32")
	}
	if positionU32 >= sizeU32 {
		panic("bit vector position out of range")
	}
	return position >> 3, 1 << (position & 0x07)
}

func (b *fixBitVector) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	size := rr.ReadSizeWithLimit(math.MaxUint16)
	sizeU16, err := safecast.Convert[uint16](size)
	if err != nil {
		return fmt.Errorf("size too large for uint16: %d", size)
	}
	b.size = sizeU16
	size = rr.CheckAvailable((size + 7) / 8)
	b.data = make([]byte, size)
	rr.ReadN(b.data)
	return rr.Err
}

func (b *fixBitVector) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteSize16(int(b.size))
	ww.WriteN(b.data)
	return ww.Err
}
