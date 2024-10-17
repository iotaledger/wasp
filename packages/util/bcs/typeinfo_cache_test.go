package bcs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestTypeInfoCacheConcurrency(t *testing.T) {
	encoderGlobalTypeInfoCache.entries.Store(&map[reflect.Type]typeInfo{})
	decoderGlobalTypeInfoCache.entries.Store(&map[reflect.Type]typeInfo{})

	type TestStruct[T any] struct {
		A T
	}

	encode := func() error {
		e := NewBytesEncoder()

		e.MustEncode(TestStruct[int32]{A: 10})
		e.MustEncode(TestStruct[string]{A: "aaa"})
		e.MustEncode(TestStruct[[]byte]{A: []byte{1, 2, 3}})
		require.NoError(t, e.Err())

		b := e.Bytes()

		d := NewBytesDecoder(b)
		Decode[TestStruct[int32]](&d.Decoder)
		Decode[TestStruct[string]](&d.Decoder)
		Decode[TestStruct[[]byte]](&d.Decoder)
		require.NoError(t, d.Err())

		encoderGlobalTypeInfoCache.entries.Store(&map[reflect.Type]typeInfo{})
		decoderGlobalTypeInfoCache.entries.Store(&map[reflect.Type]typeInfo{})

		return nil
	}

	g := errgroup.Group{}

	for i := 0; i < 1000; i++ {
		g.Go(encode)
	}

	err := g.Wait()
	require.NoError(t, err)
}
