package hashing

import (
	"fmt"
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
	h2, e := HashValueFromString(h1.String())
	if e != nil {
		t.Fatalf("error occurs")
	}
	if !h2.Equal(h1) {
		t.Fatalf("error occurs")
	}
}

func TestHashData(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashData(bytes)
	if reflect.TypeOf(nilHash) != reflect.TypeOf(*h) {
		t.Fatalf("failed to hash bytes array")
	}
}

func TestHashDataBlake2b(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashDataBlake2b(bytes)
	if reflect.TypeOf(nilHash) != reflect.TypeOf(*h) {
		t.Fatalf("failed to hash bytes array with blake 2b")
	}
}

func TestHashDataSha3(t *testing.T) {
	var bytes = []byte{0, 1, 2, 3}
	h := HashDataSha3(bytes)
	if reflect.TypeOf(nilHash) != reflect.TypeOf(*h) {
		t.Fatalf("failed to hash bytes array with sha3")
	}
}

func TestHashStrings(t *testing.T) {
	var str = []string{"kuku", "mumu", "zuzu", "rrrr"}
	h := HashStrings(str...)
	if reflect.TypeOf(nilHash) != reflect.TypeOf(*h) {
		t.Fatalf("failed to hash string array")
	}
}

func TestRandomHash(t *testing.T) {
	var src = &SampleSource{
		seed: 1,
	}
	var rnd = rand.New(src)
	h := RandomHash(rnd)
	if reflect.TypeOf(nilHash) != reflect.TypeOf(*h) {
		t.Fatalf("failed to generate random hash")
	}
}

func TestHashInList(t *testing.T) {
	var seed1 = HashStrings("alice").String()
	var seed2 = HashStrings("bob").String()
	var seed3 = HashStrings("crea").String()
	var seed4 = HashStrings("david").String()
	h1, _ := HashValueFromString(seed1)
	h2, _ := HashValueFromString(seed2)
	h3, _ := HashValueFromString(seed3)
	h4, _ := HashValueFromString(seed4)
	hashArray := []*HashValue{h1, h2, h3}
	res1 := HashInList(h1, hashArray)
	if !res1 {
		t.Fatalf("failed to check")
	}
	res2 := HashInList(h4, hashArray)
	if res2 == true {
		t.Fatalf("failed to check")
	}
}

func TestBytes(t *testing.T) {
	var bytesArray []byte
	var h1 = HashStrings("alice")
	var bytes = h1.Bytes()
	if reflect.TypeOf(bytesArray) != reflect.TypeOf(bytes) {
		t.Fatalf("failed to convert hash to bytes array")
	}
}

func TestString(t *testing.T) {
	var stringType string
	var h1 = HashStrings("alice")
	var stringified = h1.String()
	if reflect.TypeOf(stringType) != reflect.TypeOf(stringified) {
		t.Fatalf("failed to convert hash to bytes array")
	}
}

func TestEqual(t *testing.T) {
	var h1 = HashStrings("alice")
	var h2 = HashStrings("alice")
	isEqual := h1.Equal(h2)
	if !isEqual {
		t.Fatalf("failed to check")
	}
}

func TestClone(t *testing.T) {
	var h1 = HashStrings("alice")
	var h2 = h1.Clone()
	if *h1 != *h2 {
		t.Fatalf("failed to check")
	}
}
