package colored

import (
	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/stringify"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// Balances represents a collection of balances associated to their respective Color that maintains a
// deterministic order of the present Colors.
type Balances map[Color]uint64

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
		col, colorErr := ColorFromMarshalUtil(marshalUtil)
		if colorErr != nil {
			return nil, xerrors.Errorf("failed to parse Color from MarshalUtil: %w", colorErr)
		}

		// check semantic correctness (enforce ordering)
		if previousColor != nil && previousColor.Compare(&col) >= 0 {
			return nil, xerrors.Errorf("parsed Colors are not in correct order: %w", cerrors.ErrParseBytesFailed)
		}

		balance, balanceErr := marshalUtil.ReadUint64()
		if balanceErr != nil {
			return nil, xerrors.Errorf("failed to parse balance of Color %s (%v): %w", col.String(), balanceErr, cerrors.ErrParseBytesFailed)
		}
		if balance == 0 {
			return nil, xerrors.Errorf("zero balance found for color %s", col.String())
		}
		ret[col] = balance

		previousColor = &col
	}
	return ret, nil
}

// Get returns the balance of the given Color. 0 means balance is empty
func (c Balances) Get(color Color) uint64 {
	return c[color]
}

// Get returns the balance of the given Color.
func (c Balances) Set(col Color, bal uint64) Balances {
	if bal > 0 {
		c[col] = bal
	} else {
		delete(c, col)
	}
	return c
}

func (c Balances) IsEmpty() bool {
	return len(c) == 0
}

func (c Balances) Add(col Color, bal uint64) Balances {
	if bal > 0 {
		c[col] += bal
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
func (c Balances) ForEachRandomly(consumer func(col Color, bal uint64) bool) {
	for col, bal := range c {
		if !consumer(col, bal) {
			return
		}
	}
}

// ForEach calls the consumer for each element in the collection and aborts the iteration if the consumer returns false.
// Deterministic order of iteration
func (c Balances) ForEachSorted(consumer func(col Color, bal uint64) bool) {
	keys := make([]Color, 0, len(c))
	for col := range c {
		keys = append(keys, col)
	}
	Sort(keys)
	for _, col := range keys {
		if c[col] > 0 && !consumer(col, c[col]) {
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

func BalancesFromDict(d dict.Dict) (Balances, error) {
	ret := NewBalances()
	for key, value := range d {
		col, err := ColorFromBytes([]byte(key))
		if err != nil {
			return nil, err
		}
		v, err := util.Uint64From8Bytes(value)
		if err != nil {
			return nil, err
		}
		ret[col] = v
	}
	return ret, nil
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
