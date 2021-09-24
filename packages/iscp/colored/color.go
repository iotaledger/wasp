package colored

import (
	"bytes"
	"crypto/rand"
	"sort"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// Color is abstract color code used in ISCP.
// It can be mapped into specific implementations of Goshimmer or Chrysalis by calling Init
type Color [ColorLength]byte

// ColorFromBytes unmarshals a Color from a sequence of bytes.
func ColorFromBytes(colorBytes []byte) (ret Color, err error) {
	ret, err = ColorFromMarshalUtil(marshalutil.New(colorBytes))
	return
}

// ColorFromBase58EncodedString creates a Color from a base58 encoded string.
func ColorFromBase58EncodedString(base58String string) (ret Color, err error) {
	parsedBytes, err := base58.Decode(base58String)
	if err != nil {
		err = xerrors.Errorf("ColorFromBase58EncodedString (%v): %w", err, cerrors.ErrBase58DecodeFailed)
		return
	}
	col, err := ColorFromBytes(parsedBytes)
	if err != nil {
		err = xerrors.Errorf("ColorFromBase58EncodedString: %w", err)
		return
	}
	return col, nil
}

// ColorFromMarshalUtil unmarshals a Color using a MarshalUtil (for easier unmarshaling).
func ColorFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (ret Color, err error) {
	colorBytes, err := marshalUtil.ReadBytes(ColorLength)
	if err != nil {
		err = xerrors.Errorf("ColorFromMarshalUtil (%v): %w", err, cerrors.ErrParseBytesFailed)
		return
	}
	copy(ret[:], colorBytes)
	return
}

func ColorRandom() (ret Color) {
	_, err := rand.Read(ret[:])
	if err != nil {
		panic(err)
	}
	return
}

func (c *Color) Clone() (ret Color) {
	copy(ret[:], c[:])
	return
}

// Bytes marshals the Color into a sequence of bytes.
func (c *Color) Bytes() []byte {
	return c[:]
}

// Base58 returns a base58 encoded version of the Color.
func (c *Color) Base58() string {
	return base58.Encode(c.Bytes())
}

// String creates a human readable string of the Color.
func (c *Color) String() string {
	switch {
	case *c == IOTA:
		return "IOTA"
	default:
		return c.Base58()
	}
}

func (c *Color) Compare(another *Color) int {
	return bytes.Compare(c[:], another[:])
}

func Sort(arr []Color) {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Compare(&arr[j]) < 0
	})
}
