package wbf_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/util/wbf"
	"github.com/stretchr/testify/require"
)

func testCodec[V any](t *testing.T, v V) {
	vEnc := must2(wbf.Encode(v))
	vDec := must2(wbf.Decode[V](vEnc))
	require.Equal(t, v, vDec)
}

func TestDecoder(t *testing.T) {
	//var err error

	testCodec(t, BasicStruct{A: 42, B: "aaa"})
	testCodec(t, IntWithLessBytes{A: 42})
	testCodec(t, IntWithMoreBytes{A: 42})
	intV := 42
	testCodec(t, IntPtr{A: &intV})
	testCodec(t, IntOptional{A: &intV})
	testCodec(t, IntOptional{A: nil})
	testCodec(t, NestedStruct{A: 42, B: BasicStruct{A: 43, B: "aaa"}})
	testCodec(t, OptionalNestedStruct{A: 42, B: &BasicStruct{A: 43, B: "aaa"}})
	testCodec(t, OptionalNestedStruct{A: 42, B: nil})
	testCodec(t, EmbeddedStruct{BasicStruct: BasicStruct{A: 42, B: "aaa"}, C: 43})
	testCodec(t, OptionalEmbeddedStruct{BasicStruct: &BasicStruct{A: 42, B: "aaa"}, C: 43})
	testCodec(t, OptionalEmbeddedStruct{BasicStruct: nil, C: 43})
	testCodec(t, WithSlice{A: []int{42, 43}})
	testCodec(t, WithSlice{A: nil})
	testCodec(t, WithShortSlice{A: []int{42, 43}})
	testCodec(t, WithOptionalSlice{A: &[]int{42, 43}})
	testCodec(t, WithArray{A: [3]int{42, 43, 44}})
	testCodec(t, WithBigIntPtr{A: big.NewInt(42)})
	testCodec(t, WithBigIntVal{A: *big.NewInt(42)})
	testCodec(t, WithTime{A: time.Unix(12345, 6789)})
	testCodec(t, WithCustomCodec{})
	testCodec(t, WithNestedCustomCodec{A: 43, B: WithCustomCodec{}})
	testCodec(t, WithWBFOpts{A: 42})
	testCodec(t, WithWBFOptsOverride{A: 42})

	{
		v := WitUnexported{A: 42, b: 43, c: 44, D: 45}
		vEnc := must2(wbf.Encode(v))
		vDec := must2(wbf.Decode[WitUnexported](vEnc))
		require.NotEqual(t, v, vDec)
		require.Equal(t, 0, vDec.b)
		require.Equal(t, 0, vDec.D)
		vDec.b = 43
		vDec.D = 45
		require.Equal(t, v, vDec)
	}
}
