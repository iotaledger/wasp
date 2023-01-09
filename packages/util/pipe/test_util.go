package pipe

import (
	"encoding/binary"

	"github.com/iotaledger/wasp/packages/hashing"
)

type IntBased interface {
	AsInt() int
}

type Factory[E IntBased] interface {
	Create(int) E
}

type SimpleHashable int
type SimpleNothashable int

var _ Hashable = SimpleHashable(0)
var _ IntBased = SimpleHashable(0)
var _ IntBased = SimpleNothashable(0)

func (sh SimpleHashable) GetHash() hashing.HashValue {
	bin := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(bin, int64(sh))
	return hashing.HashData(bin)
}

func (sh SimpleHashable) AsInt() int {
	return int(sh)
}

func (snh SimpleNothashable) AsInt() int {
	return int(snh)
}

//--

type SimpleHashableFactory struct{}

func (_ *SimpleHashableFactory) Create(i int) SimpleHashable { return SimpleHashable(i) }

func NewSimpleHashableFactory() Factory[SimpleHashable] { return &SimpleHashableFactory{} }

type SimpleNothashableFactory struct{}

func (_ *SimpleNothashableFactory) Create(i int) SimpleNothashable { return SimpleNothashable(i) }

func NewSimpleNothashableFactory() Factory[SimpleNothashable] { return &SimpleNothashableFactory{} }

//--

func identityFunInt(index int) int {
	return index
}

func alwaysTrueFun(index int) bool {
	_ = index
	return true
}

func priorityFunMod2[E IntBased](e E) bool {
	return priorityFunMod(e, 2)
}

func priorityFunMod3[E IntBased](e E) bool {
	return priorityFunMod(e, 3)
}

func priorityFunMod[E IntBased](e E, mod int) bool {
	return e.AsInt()%mod == 0
}
