package bcs

import (
	"reflect"
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

// Checks that:
//   - encoding and decoding succeed
//   - decoded value is equal to the original
func TestCodec[V any](t *testing.T, v V, decodeInto ...V) []byte {
	vEnc, err := Marshal(&v)
	require.NoError(t, err, "%#v", v)

	var vDec V
	if len(decodeInto) == 0 {
		vDec, err = Unmarshal[V](vEnc)
	} else {
		if len(decodeInto) != 1 {
			panic("only 1 decoding destination is allowed")
		}

		vDec = decodeInto[0]
		_, err = UnmarshalInto(vEnc, &vDec)
	}

	require.NoError(t, err, "%#v", vEnc)
	require.Equal(t, v, vDec, vEnc)
	require.NotEmpty(t, vEnc)

	return vEnc
}

// Checks that:
//   - encoding and decoding succeed
//   - decoded value is equal to the original
//   - encoded value is equal to the result of reference library
func TestCodecVsRef[V any](t *testing.T, v V) []byte {
	vEnc := TestCodec(t, v)
	require.NotEmpty(t, vEnc)

	vEncExternal := lo.Must1(ref_bcs.Marshal(v))
	require.Equal(t, vEncExternal, vEnc)

	return vEnc
}

// Checks that
//   - encoding and decoding succeed
//   - decoded value is equal to the original
//   - encoded value is equal to the expected bytes
func TestCodecAndBytes[V any](t *testing.T, v V, expectedEnc []byte) {
	vEnc := TestCodec(t, v)
	require.Equal(t, expectedEnc, vEnc)
}

// Checks that
//   - encoding and decoding succeed
//   - decoded value is equal to the original
//   - encoded value is equal to the result of reference library
//   - encoded value is equal to the expected bytes
func TestCodecAndBytesVsRef[V any](t *testing.T, v V, expectedEnc []byte) {
	vEnc := TestCodecVsRef(t, v)
	require.Equal(t, expectedEnc, vEnc)
}

// Checks that encoding fails
func TestEncodeErr[V any](t *testing.T, v V, errMustContain ...string) {
	_, err := Marshal(&v)
	require.Error(t, err)

	for _, s := range errMustContain {
		require.Contains(t, err.Error(), s)
	}
}

// Checks that decoding fails
func TestDecodeErr[V any, Encoded any](t *testing.T, v Encoded, errMustContain ...string) {
	encoded, err := Marshal(&v)
	require.NoError(t, err)

	_, err = Unmarshal[V](encoded)
	require.Error(t, err)

	for _, s := range errMustContain {
		require.Contains(t, err.Error(), s)
	}
}

// Checks that:
//   - encoding and decoding succeed
//   - decoded value is NOT equal to the original
func TestCodecAsymmetric[V any](t *testing.T, v V) {
	vEnc := lo.Must1(Marshal(&v))
	vDec := lo.Must1(Unmarshal[V](vEnc))
	require.NotEqual(t, v, vDec)
}

// Returns empty value of underlaying type.
// Can be used to provide encoding destination value for TestCodec().
// In theory TestCodec() could be doing this automatically. But this might result in some of errors
// being missed (e.g. when it is not expected that value must be preset before decoding).
// For the examples, see test TestEmpty.
func Empty[V any](v V) V {
	rv := reflect.ValueOf(&v).Elem()

	if rv.Kind() != reflect.Interface {
		var empty V
		return empty
	}

	underlayingValueType := rv.Elem().Type()
	emptyUnderlayingValue := reflect.New(underlayingValueType).Elem()

	return emptyUnderlayingValue.Interface().(V)
}
