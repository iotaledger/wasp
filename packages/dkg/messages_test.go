package dkg

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

func TestInitiatorMsgSerialization(t *testing.T) {
	hash1 := hashing.PseudoRandomHash(nil).Bytes()
	pubKey1, err1 := cryptolib.PublicKeyFromBytes(hash1)
	require.NoError(t, err1)

	hash2 := hashing.PseudoRandomHash(nil).Bytes()
	pubKey2, err2 := cryptolib.PublicKeyFromBytes(hash2)
	require.NoError(t, err2)

	hash3 := hashing.PseudoRandomHash(nil).Bytes()
	pubKey3, err3 := cryptolib.PublicKeyFromBytes(hash3)
	require.NoError(t, err3)

	// Set up a random initiatorInitMsg.
	// Make sure to fill in all the fields that get serialized.
	// First we test with an empty peerPubs array.
	msg := &initiatorInitMsg{
		step:         69,
		dkgRef:       "some text",
		peerPubs:     []*cryptolib.PublicKey{},
		initiatorPub: pubKey1,
		threshold:    12321,
		timeout:      time.Duration(time.Now().UnixNano()),
		roundRetry:   time.Duration(time.Now().UnixNano() + 1),
	}
	bcs.TestCodec(t, msg)

	// Test with a 1-item peerPubs array.
	msg.peerPubs = []*cryptolib.PublicKey{pubKey2}
	bcs.TestCodec(t, msg)

	// Test with a 3-item peerPubs array.
	msg.peerPubs = []*cryptolib.PublicKey{pubKey3, pubKey2, pubKey1}
	bcs.TestCodec(t, msg)

	pubKey1 = lo.Must(cryptolib.PublicKeyFromBytes(lo.RepeatBy(cryptolib.PublicKeySize, func(i int) byte {
		return byte(i % 256)
	})))
	pubKey2 = lo.Must(cryptolib.PublicKeyFromBytes(lo.RepeatBy(cryptolib.PublicKeySize, func(i int) byte {
		return byte((i * 2) % 256)
	})))
	pubKey3 = lo.Must(cryptolib.PublicKeyFromBytes(lo.RepeatBy(cryptolib.PublicKeySize, func(i int) byte {
		return byte((i * 3) % 256)
	})))

	msg = &initiatorInitMsg{
		step:         69,
		dkgRef:       "some text",
		peerPubs:     []*cryptolib.PublicKey{},
		initiatorPub: pubKey1,
		threshold:    12321,
		timeout:      time.Duration(123),
		roundRetry:   time.Duration(456),
	}
	bcs.TestCodecAndHash(t, msg, "abb7088a6892")
	msg.peerPubs = []*cryptolib.PublicKey{pubKey2}
	bcs.TestCodecAndHash(t, msg, "f717245cc6b4")
	msg.peerPubs = []*cryptolib.PublicKey{pubKey3, pubKey2, pubKey1}
	bcs.TestCodecAndHash(t, msg, "7e77eed09664")
}
