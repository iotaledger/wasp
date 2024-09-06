// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type BitVector interface {
	SetBits(positions []int) BitVector
	AsInts() []int
	Bytes() []byte
}

func init() {
	bcs.RegisterEnumType1[BitVector, *fixBitVector]()
}

type fixBitVector struct {
	size uint16 `bcs:""`
	data []byte `bcs:""`
}

func NewFixedSizeBitVector(size uint16) BitVector {
	return &fixBitVector{size: size, data: make([]byte, (size+7)/8)}
}

func FixedSizeBitVectorFromBytes(data []byte) (BitVector, error) {
	return bcs.Unmarshal[*fixBitVector](data)
}

func (b *fixBitVector) Bytes() []byte {
	return bcs.MustMarshal(b)
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
