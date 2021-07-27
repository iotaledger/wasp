package colored

import (
	"bytes"
	"crypto/rand"
	"sort"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// region Color ////////////////////////////////////////////////////////////////////////////////////////////////////////

// ColorFromBytes unmarshals a Color from a sequence of bytes.
func ColorFromBytes(colorBytes []byte) (col Color, err error) {
	marshalUtil := marshalutil.New(colorBytes)
	if col, err = ColorFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Color from MarshalUtil: %w", err)
		return
	}
	return
}

// ColorFromBase58EncodedString creates a Color from a base58 encoded string.
func ColorFromBase58EncodedString(base58String string) (col Color, err error) {
	parsedBytes, err := base58.Decode(base58String)
	if err != nil {
		err = xerrors.Errorf("error while decoding base58 encoded Color (%v): %w", err, cerrors.ErrBase58DecodeFailed)
		return
	}

	if col, err = ColorFromBytes(parsedBytes); err != nil {
		err = xerrors.Errorf("failed to parse Color from bytes: %w", err)
		return
	}

	return
}

// ColorFromMarshalUtil unmarshals a Color using a MarshalUtil (for easier unmarshaling).
func ColorFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (col Color, err error) {
	colorBytes, err := marshalUtil.ReadBytes(ColorLength)
	if err != nil {
		err = xerrors.Errorf("failed to parse Color (%v): %w", err, cerrors.ErrParseBytesFailed)
		return
	}
	copy(col[:], colorBytes)
	return
}

func ColorRandom() (col Color) {
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

// region Balances //////////////////////////////////////////////////////////////////////////////////////////////

// Balances represents a collection of balances associated to their respective Color that maintains a
// deterministic order of the present Colors.
type Balances map[Color]uint64

var Balances1Iota = NewBalancesForIotas(1)

// NewBalances returns a new Balances. In general, it has not deterministic order
func NewBalances(bals ...map[Color]uint64) Balances {
	var b map[Color]uint64
	if len(bals) > 0 && len(bals[0]) > 0 {
		b = bals[0]
	}
	return Balances(b).Clone()
}

func NewBalancesForIotas(s uint64) Balances {
	return NewBalances().Set(IOTA, s)
}

func NewBalancesForColor(col Color, s uint64) Balances {
	return NewBalances().Set(col, s)
}

// BalancesFromBytes unmarshals Balances from a sequence of bytes.
func BalancesFromBytes(data []byte) (Balances, error) {
	marshalUtil := marshalutil.New(data)
	return BalancesFromMarshalUtil(marshalUtil)
}

// BalancesFromMarshalUtil unmarshals Balances using a MarshalUtil (for easier unmarshaling).
func BalancesFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (Balances, error) {
	balancesCount, err := marshalUtil.ReadUint32()
	if err != nil {
		return nil, xerrors.Errorf("failed to parse element count (%v): %w", err, cerrors.ErrParseBytesFailed)
	}

	var previousColor *Color
	ret := NewBalances()
	for i := uint32(0); i < balancesCount; i++ {
		color, colorErr := ColorFromMarshalUtil(marshalUtil)
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
func (c Balances) Get(color Color) uint64 {
	return c[color]
}

// Get returns the balance of the given Color.
func (c Balances) Set(color Color, bal uint64) Balances {
	if bal > 0 {
		c[color] = bal
	} else {
		delete(c, color)
	}
	return c
}

func (c Balances) IsEmpty() bool {
	return len(c) == 0
}

func (c Balances) Add(color Color, bal uint64) Balances {
	if bal > 0 {
		c[color] += bal
	}
	return c
}

// SubNoOverflow securely subtracts amount from color balance. Set to 0 is subtracted amount > existing
func (c Balances) SubNoOverflow(col Color, bal uint64) Balances {
	if bal == 0 {
		return c
	}
	if bal >= c[col] {
		c[col] = 0
	} else {
		c[col] -= bal
	}
	if c[col] == 0 {
		delete(c, col)
	}
	return c
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Non-deterministic order of iteration
func (c Balances) ForEachRandomly(consumer func(color Color, balance uint64) bool) {
	for col, bal := range c {
		if bal > 0 && !consumer(col, bal) {
			return
		}
	}
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Deterministic order of iteration
func (c Balances) ForEachSorted(consumer func(color Color, balance uint64) bool) {
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

// Clone returns a copy of the Balances.
func (c Balances) Clone() Balances {
	ret := make(Balances)
	for col, bal := range c {
		if bal > 0 {
			ret[col] = bal
		}
	}
	return ret
}

// Bytes returns a marshaled version of the Balances.
func (c Balances) Bytes() []byte {
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

// String returns a human readable version of the Balances.
func (c Balances) String() string {
	structBuilder := stringify.StructBuilder("iscp.Balances")
	c.ForEachSorted(func(color Color, balance uint64) bool {
		structBuilder.AddField(stringify.StructField(color.String(), balance))
		return true
	})
	return structBuilder.String()
}

func (c Balances) Equals(another Balances) bool {
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

func (c Balances) AddAll(another Balances) {
	another.ForEachRandomly(func(col Color, bal uint64) bool {
		c.Add(col, bal)
		return true
	})
}

// Diff returns difference between two Balances color-by-color
func (c Balances) Diff(another Balances) map[Color]int64 {
	ret := make(map[Color]int64)
	for col := range allColors(c, another) {
		cBal := c[col]
		aBal := another[col]
		switch {
		case cBal > aBal:
			ret[col] = int64(cBal - aBal)
		case cBal < aBal:
			ret[col] = -int64(aBal - cBal)
		}
	}
	return ret
}

func allColors(bals ...Balances) map[Color]bool {
	ret := make(map[Color]bool)
	for _, b := range bals {
		b.ForEachRandomly(func(col Color, bal uint64) bool {
			ret[col] = true
			return true
		})
	}
	return ret
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
