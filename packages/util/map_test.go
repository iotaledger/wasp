package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
)

func TestSome(t *testing.T) {
	t.Skip("Just a performance test.")
	count := 100000
	m1 := map[string]*cryptolib.KeyPair{}
	m2 := map[string]*cryptolib.KeyPair{}
	m3 := map[cryptolib.PublicKeyKey]*cryptolib.KeyPair{}
	k := make([]*cryptolib.KeyPair, count)
	k1 := make([]string, count)
	k2 := make([]string, count)
	k3 := make([]cryptolib.PublicKeyKey, count)
	for i := 0; i < count; i++ {
		kp := cryptolib.NewKeyPair()
		asKey := kp.GetPublicKey().AsKey()
		asStr := kp.GetPublicKey().String()
		m1[asStr] = kp
		m2[string(asKey[:])] = kp
		m3[asKey] = kp
		k[i] = kp
		k1[i] = asStr
		k2[i] = string(asKey[:])
		k3[i] = asKey
	}

	now := time.Now()
	for i := range k {
		require.Equal(t, k[i], m1[k1[i]])
	}
	d1 := time.Since(now)
	now = time.Now()
	for i := range k {
		require.Equal(t, k[i], m2[k2[i]])
	}
	d2 := time.Since(now)
	now = time.Now()
	for i := range k {
		require.Equal(t, k[i], m3[k3[i]])
	}
	d3 := time.Since(now)
	t.Logf("m1=%v, m2=%v, m3=%v", d1, d2, d3)
}
