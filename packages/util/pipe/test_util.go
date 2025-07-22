package pipe

import (
	"encoding/binary"

	"github.com/iotaledger/wasp/v2/packages/hashing"
)

type IntConvertible interface {
	AsInt() int
}

type Factory[E IntConvertible] interface {
	Create(int) E
}

type SimpleHashable int

type SimpleNothashable int

var (
	_ Hashable                   = SimpleHashable(0)
	_ IntConvertible             = SimpleHashable(0)
	_ IntConvertible             = SimpleNothashable(0)
	_ Factory[SimpleHashable]    = &SimpleHashableFactory{}
	_ Factory[SimpleNothashable] = &SimpleNothashableFactory{}
)

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

func (*SimpleHashableFactory) Create(i int) SimpleHashable { return SimpleHashable(i) }

func NewSimpleHashableFactory() Factory[SimpleHashable] { return &SimpleHashableFactory{} }

type SimpleNothashableFactory struct{}

func (*SimpleNothashableFactory) Create(i int) SimpleNothashable { return SimpleNothashable(i) }

func NewSimpleNothashableFactory() Factory[SimpleNothashable] { return &SimpleNothashableFactory{} }

//--

func identityFunInt(index int) int {
	return index
}

func alwaysTrueFun(index int) bool {
	_ = index
	return true
}

func priorityFunMod2[E IntConvertible](e E) bool {
	return priorityFunMod(e, 2)
}

func priorityFunMod3[E IntConvertible](e E) bool {
	return priorityFunMod(e, 3)
}

func priorityFunMod[E IntConvertible](e E, mod int) bool {
	return e.AsInt()%mod == 0
}
