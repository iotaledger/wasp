package util

import (
	"fmt"
	"math/big"
)

func ToBigInt(i interface{}) *big.Int {
	switch it := i.(type) {
	case *big.Int:
		return it
	case int:
		return big.NewInt(int64(it))
	case uint64:
		return new(big.Int).SetUint64(it)
	case uint32:
		return big.NewInt(int64(it))
	case uint16:
		return big.NewInt(int64(it))
	case uint8:
		return big.NewInt(int64(it))
	case int64:
		return big.NewInt(it)
	case int32:
		return big.NewInt(int64(it))
	case int16:
		return big.NewInt(int64(it))
	case int8:
		return big.NewInt(int64(it))
	}
	panic(fmt.Sprintf("ToBigInt: type %T not supported", i))
}
