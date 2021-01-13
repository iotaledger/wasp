package hashing

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"reflect"
	"testing"
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
	var h1 = HashStrings("test string")
	h2, e := HashValueFromBase58(h1.String())
	if e != nil {
		t.Fatalf("error occurs")
	}
	if h2 != h1 {
		t.Fatalf("error occurs")
	}
}

func TestHashData(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashData(bytes)
	if reflect.TypeOf(nilHash) != reflect.TypeOf(h) {
		t.Fatalf("failed to hash bytes array")
	}
}

func TestHashStrings(t *testing.T) {
	var str = []string{"kuku", "mumu", "zuzu", "rrrr"}
	h := HashStrings(str...)
	require.EqualValues(t, reflect.TypeOf(nilHash), reflect.TypeOf(h))
}

func TestRandomHash(t *testing.T) {
	var src = &SampleSource{
		seed: 1,
	}
	var rnd = rand.New(src)
	h := RandomHash(rnd)
	require.EqualValues(t, reflect.TypeOf(nilHash), reflect.TypeOf(*h))
}

func TestString(t *testing.T) {
	var stringType string
	var h1 = HashStrings("alice")
	var stringified = h1.String()
	require.EqualValues(t, reflect.TypeOf(stringType), reflect.TypeOf(stringified))
	require.EqualValues(t, h1.String(), (&h1).String())
	require.EqualValues(t, h1.Short(), (&h1).Short())
	require.EqualValues(t, h1.Shortest(), (&h1).Shortest())
}
