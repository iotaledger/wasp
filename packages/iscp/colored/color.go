package colored

import (
	"bytes"
	"crypto/rand"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// Color is abstract color code used in ISCP.
// It can be mapped into specific implementations of Goshimmer or Chrysalis by calling Init
type Color []byte
type ColorKey string

// colorLength represents the length of a Color (amount of bytes).
var colorLength int

// IOTA is the zero value of the Color and represents uncolored tokens.
var IOTA = Color(ledgerstate.ColorIOTA[:])

func ColorLength() int {
	return colorLength
}

// Init initializes module with the specific length of the color code and specific code used as IOTA color code
func Init(colLen int, iotaColor Color) {
	if colorLength != 0 {
		panic("color20.init called twice")
	}
	colorLength = colLen
	IOTA = iotaColor
}

func NewColor(key ...ColorKey) (Color, error) {
	ret := make(Color, colorLength)
	if len(key) == 0 {
		return ret, nil
	}
	if len(key) > 0 {
		if len(key[0]) != colorLength {
			return nil, xerrors.Errorf("ColorFromBytes: %d bytes expected", colorLength)
		}
	}
	copy(ret, key[0])
	return ret, nil
}

// ColorFromBytes unmarshals a Color from a sequence of bytes.
func ColorFromBytes(colorBytes []byte) (col Color, err error) {
	return NewColor(ColorKey(colorBytes))
}

// ColorFromBase58EncodedString creates a Color from a base58 encoded string.
func ColorFromBase58EncodedString(base58String string) (Color, error) {
	parsedBytes, err := base58.Decode(base58String)
	if err != nil {
		return nil, xerrors.Errorf("ColorFromBase58EncodedString (%v): %w", err, cerrors.ErrBase58DecodeFailed)
	}

	col, err := ColorFromBytes(parsedBytes)
	if err != nil {
		return nil, xerrors.Errorf("ColorFromBase58EncodedString: %w", err)
	}
	return col, nil
}

// ColorFromMarshalUtil unmarshals a Color using a MarshalUtil (for easier unmarshaling).
func ColorFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (Color, error) {
	colorBytes, err := marshalUtil.ReadBytes(colorLength)
	if err != nil {
		return nil, xerrors.Errorf("ColorFromMarshalUtil (%v): %w", err, cerrors.ErrParseBytesFailed)
	}
	return colorBytes, nil
}

func ColorRandom() Color {
	ret, err := NewColor()
	if err != nil {
		panic(err)
	}
	_, err = rand.Read(ret[:])
	if err != nil {
		panic(err)
	}
	return ret
}

// Bytes marshals the Color into a sequence of bytes.
func (c Color) Bytes() []byte {
	return c[:]
}

// Base58 returns a base58 encoded version of the Color.
func (c Color) Base58() string {
	return base58.Encode(c.Bytes())
}

// String creates a human readable string of the Color.
func (c Color) String() string {
	switch {
	case c.Compare(IOTA) == 0:
		return "IOTA"
	default:
		return c.Base58()
	}
}

func (c Color) AsKey() ColorKey {
	return ColorKey(c)
}

// Compare offers a comparator for Colors which returns -1 if otherColor is bigger, 1 if it is smaller and 0 if they are
// the same.
func (c Color) Compare(otherColor Color) int {
	return bytes.Compare(c[:], otherColor[:])
}
