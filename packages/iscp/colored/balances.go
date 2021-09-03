package colored

import (
	"sort"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
	"golang.org/x/xerrors"
)

// Balances represents a collection of balances associated to their respective Color that maintains a
// deterministic order of the present Colors.
type Balances map[ColorKey]uint64

var Balances1Iota = NewBalancesForIotas(1)

// NewBalances returns a new Balances. In general, it has not deterministic order
func NewBalances() Balances {
	return make(Balances)
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
		ret[color.AsKey()] = balance

		previousColor = &color
	}
	return ret, nil
}

// Get returns the balance of the given Color. 0 means balance is empty
func (c Balances) Get(color Color) uint64 {
	return c[color.AsKey()]
}

// Get returns the balance of the given Color.
func (c Balances) Set(color Color, bal uint64) Balances {
	k := color.AsKey()
	if bal > 0 {
		c[k] = bal
	} else {
		delete(c, k)
	}
	return c
}

func (c Balances) IsEmpty() bool {
	return len(c) == 0
}

func (c Balances) Add(col Color, bal uint64) Balances {
	if bal > 0 {
		c[col.AsKey()] += bal
	}
	return c
}

// SubNoOverflow securely subtracts amount from color balance. Set to 0 is subtracted amount > existing
func (c Balances) SubNoOverflow(col Color, bal uint64) Balances {
	if bal == 0 {
		return c
	}
	k := col.AsKey()
	if bal >= c[k] {
		c[k] = 0
	} else {
		c[k] -= bal
	}
	if c[k] == 0 {
		delete(c, col.AsKey())
	}
	return c
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Non-deterministic order of iteration
func (c Balances) ForEachRandomly(consumer func(color Color, balance uint64) bool) {
	for col, bal := range c {
		if bal > 0 && !consumer(Color(col), bal) {
			return
		}
	}
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Deterministic order of iteration
func (c Balances) ForEachSorted(consumer func(color Color, balance uint64) bool) {
	keys := make([]string, 0, len(c))
	for col := range c {
		keys = append(keys, string(col))
	}
	sort.Strings(keys)
	for _, col := range keys {
		bal := c[ColorKey(col)]
		if bal > 0 && !consumer(Color(col), bal) {
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
func (c Balances) Diff(another Balances) map[ColorKey]int64 {
	ret := make(map[ColorKey]int64)
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

func allColors(bals ...Balances) map[ColorKey]bool {
	ret := make(map[ColorKey]bool)
	for _, b := range bals {
		b.ForEachRandomly(func(col Color, bal uint64) bool {
			ret[col.AsKey()] = true
			return true
		})
	}
	return ret
}
