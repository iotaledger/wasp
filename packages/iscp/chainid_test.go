package iscp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainID(t *testing.T) {
	chid := RandomChainID()

	chidStr := chid.String()
	t.Logf("chidStr = %s", chidStr)

	chidHex := chid.Hex()
	t.Logf("chidHex = %s", chidHex)

	chidback, err := ChainIDFromBytes(chid.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)

	chidback, err = ChainIDFromHex(chidHex)
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)
}
