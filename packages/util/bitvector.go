// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

type BitVector interface {
	SetBits(positions []int) BitVector
	AsInts() []int
	Bytes() []byte
}

type fixBitVector struct {
	size int
	data []byte
}

func NewFixedSizeBitVector(size int) BitVector {
	return &fixBitVector{size: size, data: make([]byte, (size-1)/8+1)}
}

func NewFixedSizeBitVectorFromMarshalUtil(mu *marshalutil.MarshalUtil) (BitVector, error) {
	size, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	byteSize := (int(size)-1)/8 + 1
	data, err := mu.ReadBytes(byteSize)
	if err != nil {
		return nil, err
	}
	return &fixBitVector{size: int(size), data: data}, nil
}

func (b *fixBitVector) SetBits(positions []int) BitVector {
	for _, p := range positions {
		bytePos, bitMask := b.bitMask(p)
		b.data[bytePos] |= bitMask
	}
	return b
}

func (b *fixBitVector) Bytes() []byte {
	return marshalutil.New().WriteUint16(uint16(b.size)).WriteBytes(b.data).Bytes()
}

func (b *fixBitVector) AsInts() []int {
	ints := []int{}
	for i := 0; i < b.size; i++ {
		bytePos, bitMask := b.bitMask(i)
		if b.data[bytePos]&bitMask != 0 {
			ints = append(ints, i)
		}
	}
	return ints
}

func (b *fixBitVector) bitMask(position int) (int, byte) {
	var bitMask byte = 1
	bitMask <<= position % 8
	bytePos := position / 8
	return bytePos, bitMask
}
