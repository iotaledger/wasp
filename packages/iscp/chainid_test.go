package iscp

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/stretchr/testify/assert"
)

func TestChainID(t *testing.T) {
	chid := RandomChainID()

	chid58 := chid.Bech32(iotago.PrefixTestnet)
	t.Logf("chid58 = %s", chid58)

	chidString := chid.String()
	t.Logf("chidString = %s", chidString)

	chidback, err := ChainIDFromBytes(chid.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)

	chidback, err = ChainIDFromBase58(chid58)
	assert.NoError(t, err)
	assert.EqualValues(t, chidback, chid)
}
