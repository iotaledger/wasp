package codec

import "math/big"

var (
	Int8      = NewCodecFromBCS[int8]()
	Uint8     = NewCodecFromBCS[uint8]()
	Int16     = NewCodecFromBCS[int16]()
	Uint16    = NewCodecFromBCS[uint16]()
	Int32     = NewCodecFromBCS[int32]()
	Uint32    = NewCodecFromBCS[uint32]()
	Int64     = NewCodecFromBCS[int64]()
	Uint64    = NewCodecFromBCS[uint64]()
	BigIntAbs = NewCodecFromBCS[*big.Int]()
)
