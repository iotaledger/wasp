package tcrypto

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewRandomDKShare(t *testing.T) {
	ks, err := NewRndDKShare(67, 100, 5)
	assert.Equal(t, err, nil)

	assert.Equal(t, ks != nil, true)
	assert.Equal(t, ks.N, uint16(100))
	assert.Equal(t, ks.T, uint16(67))
	assert.Equal(t, ks.Index, uint16(5))
	assert.Equal(t, len(ks.PriShares), 100)
	assert.Equal(t, ks.Aggregated, false)
	assert.Equal(t, ks.Committed, false)

	_, err = NewRndDKShare(1, 1, 0)
	assert.Equal(t, err, nil)

	_, err = NewRndDKShare(5, 4, 0)
	assert.Equal(t, err != nil, true)

	_, err = NewRndDKShare(4, 5, 6)
	assert.Equal(t, err != nil, true)
}
