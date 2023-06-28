// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"io"
	"math"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type BitVector interface {
	SetBits(positions []int) BitVector
	AsInts() []int
	Bytes() []byte
	Read(r io.Reader) error
	Write(w io.Writer) error
}

type fixBitVector struct {
	size uint16
	data []byte
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
	if uint32(position) >= uint32(b.size) {
		panic("bit vector position out of range")
	}
	return position >> 3, 1 << (position & 0x07)
}

func (b *fixBitVector) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	size := rr.ReadSizeWithLimit(math.MaxUint16)
	b.size = uint16(size)
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
