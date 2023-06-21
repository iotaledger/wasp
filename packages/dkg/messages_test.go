package dkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestInitiatorMsgSerialization(t *testing.T) {
	hash1 := hashing.PseudoRandomHash(nil).Bytes()
	pubKey1, err := cryptolib.PublicKeyFromBytes(hash1)
	require.NoError(t, err)
	hash2 := hashing.PseudoRandomHash(nil).Bytes()
	pubKey2, err := cryptolib.PublicKeyFromBytes(hash2)
	require.NoError(t, err)
	hash3 := hashing.PseudoRandomHash(nil).Bytes()
	pubKey3, err := cryptolib.PublicKeyFromBytes(hash3)
	require.NoError(t, err)

	// set up a random initiatorInitMsg
	// make sure to fill in all the fields that get serialized
	// first test with an empty peerPubs array
	msg := &initiatorInitMsg{
		step:         69,
		dkgRef:       "some text",
		peerPubs:     []*cryptolib.PublicKey{},
		initiatorPub: pubKey1,
		threshold:    12321,
		timeout:      time.Duration(time.Now().UnixNano()),
		roundRetry:   time.Duration(time.Now().UnixNano() + 1),
	}
	rwutil.SerializationTest(t, msg, new(initiatorInitMsg))

	//  test with a 1-item peerPubs array
	msg.peerPubs = []*cryptolib.PublicKey{pubKey2}
	rwutil.SerializationTest(t, msg, new(initiatorInitMsg))

	//  test with a 3-item peerPubs array
	msg.peerPubs = []*cryptolib.PublicKey{pubKey3, pubKey2, pubKey1}
	rwutil.SerializationTest(t, msg, new(initiatorInitMsg))
}
