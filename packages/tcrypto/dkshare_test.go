package tcrypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomDKShare(t *testing.T) {
	ks, err := NewRndDKShare(67, 100, 5)
	assert.NoError(t, err)

	assert.NotNil(t, ks)
	assert.Equal(t, ks.N, uint16(100))
	assert.Equal(t, ks.T, uint16(67))
	assert.Equal(t, ks.Index, uint16(5))
	assert.Equal(t, len(ks.PriShares), 100)
	assert.Equal(t, ks.Aggregated, false)
	assert.Equal(t, ks.Committed, false)

	_, err = NewRndDKShare(2, 2, 0)
	assert.NoError(t, err)

	_, err = NewRndDKShare(5, 4, 0)
	assert.Error(t, err)

	_, err = NewRndDKShare(4, 5, 6)
	assert.Error(t, err)
}
