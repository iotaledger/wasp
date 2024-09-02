package bcs

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
)

func init() {
	AddCustomEncoder(func(e *Encoder, v big.Int) error {
		return EncodeUint128(&v, e)
	})

	AddCustomDecoder(func(d *Decoder) (big.Int, error) {
		v, err := DecodeUint128(d)
		if err != nil {
			return big.Int{}, err
		}

		return *v, nil
	})
}

func EncodeUint128(v *big.Int, w io.Writer) error {
	if err := checkUint128(v); err != nil {
		return fmt.Errorf("checking Uint128 validity: %w", err)
	}

	hi, lo := bigIntToUint128(v)
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, lo)
	binary.LittleEndian.PutUint64(b[8:], hi)

	_, err := w.Write(b)
	return err
}

func bigIntToUint128(bigI *big.Int) (uint64, uint64) {
	r := make([]byte, 0, 16)
	bs := bigI.Bytes()

	for i := 0; i+len(bs) < 16; i++ {
		r = append(r, 0)
	}

	r = append(r, bs...)

	hi := binary.BigEndian.Uint64(r[0:8])
	lo := binary.BigEndian.Uint64(r[8:16])

	return hi, lo
}

var maxU128 = (&big.Int{}).Lsh(big.NewInt(1), 128)

func checkUint128(bigI *big.Int) error {
	if bigI.Sign() < 0 {
		return fmt.Errorf("%s is negative", bigI.String())
	}

	if bigI.Cmp(maxU128) >= 0 {
		return fmt.Errorf("%s is greater than max Uint128", bigI.String())
	}

	return nil
}

func DecodeUint128(r io.Reader) (*big.Int, error) {
	buf := make([]byte, 16)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 16 {
		return nil, fmt.Errorf("got unexpected number of bytes for Uint128: expected %v, got %v", len(buf), n)
	}

	lo := binary.LittleEndian.Uint64(buf[0:8])
	hi := binary.LittleEndian.Uint64(buf[8:16])

	loBig := uint64ToBigInt(lo)
	hiBig := uint64ToBigInt(hi)

	hiBig = hiBig.Lsh(hiBig, 64)

	return hiBig.Add(hiBig, loBig), nil
}

// 63 ones
const ones63 uint64 = (1 << 63) - 1

// 1 << 63
var oneLsh63 = big.NewInt(0).Lsh(big.NewInt(1), 63)

func uint64ToBigInt(i uint64) *big.Int {
	r := big.NewInt(int64(i & ones63))
	if i > ones63 {
		r = r.Add(r, oneLsh63)
	}
	return r
}
