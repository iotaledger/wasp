package color

import (
	"bytes"
	"crypto/rand"
	"sort"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// region Color ////////////////////////////////////////////////////////////////////////////////////////////////////////

// IOTA is the zero value of the Color and represents uncolored tokens.
var IOTA = Color{}

// Mint represents a placeholder Color that indicates that tokens should be "colored" in their Output.
var Mint = Color{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

// ColorLength represents the length of a Color (amount of bytes).
const ColorLength = ledgerstate.ColorLength

// Color represents a marker that is associated to a token balance and that can give tokens a certain "meaning".
type Color ledgerstate.Color

// FromBytes unmarshals a Color from a sequence of bytes.
func FromBytes(colorBytes []byte) (col Color, err error) {
	marshalUtil := marshalutil.New(colorBytes)
	if col, err = FromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Color from MarshalUtil: %w", err)
		return
	}
	return
}

// FromBase58EncodedString creates a Color from a base58 encoded string.
func FromBase58EncodedString(base58String string) (col Color, err error) {
	parsedBytes, err := base58.Decode(base58String)
	if err != nil {
		err = xerrors.Errorf("error while decoding base58 encoded Color (%v): %w", err, cerrors.ErrBase58DecodeFailed)
		return
	}

	if col, err = FromBytes(parsedBytes); err != nil {
		err = xerrors.Errorf("failed to parse Color from bytes: %w", err)
		return
	}

	return
}

// FromMarshalUtil unmarshals a Color using a MarshalUtil (for easier unmarshaling).
func FromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (col Color, err error) {
	colorBytes, err := marshalUtil.ReadBytes(ColorLength)
	if err != nil {
		err = xerrors.Errorf("failed to parse Color (%v): %w", err, cerrors.ErrParseBytesFailed)
		return
	}
	copy(col[:], colorBytes)
	return
}

func Random() (col Color) {
	_, err := rand.Read(col[:])
	if err != nil {
		panic(err)
	}
	return
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
	switch c {
	case IOTA:
		return "IOTA"
	case Mint:
		return "MINT"
	default:
		return c.Base58()
	}
}

// Compare offers a comparator for Colors which returns -1 if otherColor is bigger, 1 if it is smaller and 0 if they are
// the same.
func (c Color) Compare(otherColor Color) int {
	return bytes.Compare(c[:], otherColor[:])
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ColoredBalances //////////////////////////////////////////////////////////////////////////////////////////////

// ColoredBalances represents a collection of balances associated to their respective Color that maintains a
// deterministic order of the present Colors.
type ColoredBalances map[Color]uint64

// NewBalances returns a new ColoredBalances. In general, it has not deterministic order
func NewBalances(bals map[Color]uint64) ColoredBalances {
	return ColoredBalances(bals).Clone()
}

func BalancesFromGoshimmer(cb *ledgerstate.ColoredBalances) ColoredBalances {
	ret := NewBalances(nil)
	if cb != nil {
		cb.ForEach(func(color ledgerstate.Color, balance uint64) bool {
			ret.Set(Color(color), balance)
			return true
		})
	}
	return ret
}

func BalancesFromIotas(s uint64) ColoredBalances {
	return NewBalances(nil).Set(IOTA, s)
}

// BalancesFromBytes unmarshals ColoredBalances from a sequence of bytes.
func BalancesFromBytes(data []byte) (ColoredBalances, error) {
	marshalUtil := marshalutil.New(data)
	return BalancesFromMarshalUtil(marshalUtil)
}

// BalancesFromMarshalUtil unmarshals ColoredBalances using a MarshalUtil (for easier unmarshaling).
func BalancesFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (ColoredBalances, error) {
	balancesCount, err := marshalUtil.ReadUint32()
	if err != nil {
		return nil, xerrors.Errorf("failed to parse element count (%v): %w", err, cerrors.ErrParseBytesFailed)
	}
	if balancesCount == 0 {
		return nil, xerrors.Errorf("empty balances in output")
	}

	var previousColor *Color
	ret := NewBalances(nil)
	for i := uint32(0); i < balancesCount; i++ {
		color, colorErr := FromMarshalUtil(marshalUtil)
		if colorErr != nil {
			return nil, xerrors.Errorf("failed to parse Color from MarshalUtil: %w", colorErr)
		}

		// check semantic correctness (ensure ordering)
		if previousColor != nil && previousColor.Compare(color) >= 0 {
			return nil, xerrors.Errorf("parsed Colors are not in correct order: %w", cerrors.ErrParseBytesFailed)
		}

		balance, balanceErr := marshalUtil.ReadUint64()
		if balanceErr != nil {
			return nil, xerrors.Errorf("failed to parse balance of Color %s (%v): %w", color.String(), balanceErr, cerrors.ErrParseBytesFailed)
		}
		if balance == 0 {
			return nil, xerrors.Errorf("zero balance found for color %s", color.String())
		}
		ret[color] = balance

		previousColor = &color
	}
	return ret, nil
}

// Get returns the balance of the given Color. 0 means balance is empty
func (c ColoredBalances) Get(color Color) uint64 {
	return c[color]
}

// Get returns the balance of the given Color.
func (c ColoredBalances) Set(color Color, bal uint64) ColoredBalances {
	if bal > 0 {
		c[color] = bal
	} else {
		delete(c, color)
	}
	return c
}

func (c ColoredBalances) Add(color Color, bal uint64) ColoredBalances {
	if bal > 0 {
		c[color] += bal
	}
	return c
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Non-deterministic order of iteration
func (c ColoredBalances) ForEachRandomly(consumer func(color Color, balance uint64) bool) {
	for col, bal := range c {
		if bal > 0 && !consumer(col, bal) {
			return
		}
	}
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Deterministic order of iteration
func (c ColoredBalances) ForEachSorted(consumer func(color Color, balance uint64) bool) {
	keys := make([]Color, 0, len(c))
	for col := range c {
		keys = append(keys, col)
	}
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})
	for _, col := range keys {
		bal := c[col]
		if bal > 0 && !consumer(col, bal) {
			return
		}
	}
}

// Clone returns a copy of the ColoredBalances.
func (c ColoredBalances) Clone() ColoredBalances {
	ret := make(ColoredBalances)
	for col, bal := range c {
		if bal > 0 {
			ret[col] = bal
		}
	}
	return ret
}

// Bytes returns a marshaled version of the ColoredBalances.
func (c ColoredBalances) Bytes() []byte {
	marshalUtil := marshalutil.New()
	marshalUtil.WriteUint32(uint32(len(c)))
	c.ForEachSorted(func(col Color, bal uint64) bool {
		if bal > 0 {
			marshalUtil.WriteBytes(col.Bytes())
			marshalUtil.WriteUint64(bal)
		}
		return true
	})
	return marshalUtil.Bytes()
}

// String returns a human readable version of the ColoredBalances.
func (c ColoredBalances) String() string {
	structBuilder := stringify.StructBuilder("iscp.ColoredBalances")
	c.ForEachSorted(func(color Color, balance uint64) bool {
		structBuilder.AddField(stringify.StructField(color.String(), balance))
		return true
	})
	return structBuilder.String()
}

func (c ColoredBalances) Equals(another ColoredBalances) bool {
	if len(c) != len(another) {
		return false
	}
	ret := true
	c.ForEachRandomly(func(col Color, bal uint64) bool {
		if c.Get(col) != another.Get(col) {
			ret = false
			return false
		}
		return true
	})
	return ret
}

func (c ColoredBalances) AddAll(another ColoredBalances) {
	another.ForEachRandomly(func(col Color, bal uint64) bool {
		c.Add(col, bal)
		return true
	})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
