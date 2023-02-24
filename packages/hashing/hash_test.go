package hashing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type SampleSource struct {
	seed int64
}

func (s *SampleSource) Int63() int64 {
	return s.seed
}

func (s *SampleSource) Seed(seed int64) {
	fmt.Println(s)
}

func TestHashValueFromString(t *testing.T) {
	h1 := HashStrings("test string")
	h2, e := HashValueFromHex(h1.String())
	if e != nil {
		t.Fatal("error occurs")
	}
	if h2 != h1 {
		t.Fatal("error occurs")
	}
}

func TestHashData(t *testing.T) {
	bytes := []byte{0, 1, 2, 3}
	h := HashData(bytes)
	if reflect.TypeOf(NilHash) != reflect.TypeOf(h) {
		t.Fatal("failed to hash bytes array")
	}
}

func TestHashStrings(t *testing.T) {
	str := []string{"kuku", "mumu", "zuzu", "rrrr"}
	h := HashStrings(str...)
	require.EqualValues(t, reflect.TypeOf(NilHash), reflect.TypeOf(h))
}

func TestRandomHash(t *testing.T) {
	src := &SampleSource{
		seed: 1,
	}
	rnd := rand.New(src)
	h := PseudoRandomHash(rnd)
	require.EqualValues(t, reflect.TypeOf(NilHash), reflect.TypeOf(h))
}

func TestString(t *testing.T) {
	var stringType string
	h1 := HashStrings("alice")
	stringified := h1.String()
	require.EqualValues(t, reflect.TypeOf(stringType), reflect.TypeOf(stringified))
	require.EqualValues(t, h1.String(), (&h1).String())
}

func TestSha3(t *testing.T) {
	data := []byte("data-data-data-data-data-data-data-data-data")
	HashSha3(data, data, data)
}
