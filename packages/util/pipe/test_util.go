package pipe

import (
	"encoding/binary"

	"github.com/iotaledger/wasp/packages/hashing"
)

type SimpleHashable int

var _ Hashable = SimpleHashable(0)

func (sh SimpleHashable) GetHash() hashing.HashValue {
	bin := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(bin, int64(sh))
	return hashing.HashData(bin)
}

//--

func identityFunInt(index int) int {
	return index
}

func alwaysTrueFun(index int) bool {
	_ = index
	return true
}

func priorityFunMod2(e SimpleHashable) bool {
	return priorityFunMod(e, 2)
}

func priorityFunMod3(e SimpleHashable) bool {
	return priorityFunMod(e, 3)
}

func priorityFunMod(e SimpleHashable, mod SimpleHashable) bool {
	return e%mod == 0
}
