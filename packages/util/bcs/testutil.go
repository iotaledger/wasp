package bcs

import (
	"testing"

	ref_bcs "github.com/fardream/go-bcs/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

// Checks that:
//   - encoding and decoding succeed
//   - decoded value is equal to the original
//   - encoded value is equal to the result of reference library
func TestCodec[V any](t *testing.T, v V) []byte {
	vEnc := TestCodecNoRef(t, v)
	require.NotEmpty(t, vEnc)

	vEncExternal := lo.Must1(ref_bcs.Marshal(v))
	require.Equal(t, vEncExternal, vEnc)

	return vEnc
}

// Checks that:
//   - encoding and decoding succeed
//   - decoded value is equal to the original
func TestCodecNoRef[V any](t *testing.T, v V) []byte {
	vEnc := lo.Must1(Marshal(&v))
	vDec := lo.Must1(Unmarshal[V](vEnc))
	require.Equal(t, v, vDec)
	require.NotEmpty(t, vEnc)

	return vEnc
}

// Checks that
//   - encoding and decoding succeed
//   - decoded value is equal to the original
//   - encoded value is equal to the result of reference library
//   - encoded value is equal to the expected bytes
func TestCodecAndBytes[V any](t *testing.T, v V, expectedEnc []byte) {
	vEnc := TestCodec(t, v)
	require.Equal(t, expectedEnc, vEnc)
}

// Checks that
//   - encoding and decoding succeed
//   - decoded value is equal to the original
//   - encoded value is equal to the expected bytes
func TestCodecAndBytesNoRef[V any](t *testing.T, v V, expectedEnc []byte) {
	vEnc := TestCodecNoRef(t, v)
	require.Equal(t, expectedEnc, vEnc)
}

// Checks that encoding fails
func TestEncodeErr[V any](t *testing.T, v V) {
	_, err := Marshal(&v)
	require.Error(t, err)
}

// Checks that:
//   - encoding and decoding succeed
//   - decoded value is NOT equal to the original
func TestCodecAsymmetric[V any](t *testing.T, v V) {
	vEnc := lo.Must1(Marshal(&v))
	vDec := lo.Must1(Unmarshal[V](vEnc))
	require.NotEqual(t, v, vDec)
}
