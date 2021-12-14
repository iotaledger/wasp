package iscp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainID(t *testing.T) {
	chid := RandomChainID()

	chidStr := chid.String()
	t.Logf("chidStr = %s", chidStr)

	chidString := chid.String()
	t.Logf("chidString = %s", chidString)

	chidback, err := ChainIDFromBytes(chid.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)

	chidback, err = ChainIDFromBase58(chidStr)
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)
}
