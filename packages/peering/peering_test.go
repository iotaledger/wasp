package peering_test

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/txstream/chopper"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/stretchr/testify/require"
)

func TestPeerMessageCodec(t *testing.T) {
	var err error
	var src, dst *peering.PeerMessage
	nodeIdentity := ed25519.GenerateKeyPair()
	src = &peering.PeerMessage{
		PeeringID:   peering.RandomPeeringID(),
		SenderIndex: uint16(123),
		Timestamp:   time.Now().UnixNano(),
		MsgType:     peering.FirstUserMsgCode + 17,
		MsgData:     []byte{1, 2, 3, 4, 5},
	}
	var bin []byte
	bin, err = src.Bytes(&nodeIdentity)
	require.Nil(t, err)
	require.NotNil(t, bin)
	dst, err = peering.NewPeerMessageFromBytes(bin, &nodeIdentity.PublicKey)
	require.Nil(t, err)
	require.NotNil(t, dst)
	require.EqualValues(t, src.PeeringID, dst.PeeringID)
	require.Equal(t, src.SenderIndex, dst.SenderIndex)
	require.Equal(t, src.Timestamp, dst.Timestamp)
	require.Equal(t, src.MsgType, dst.MsgType)
	require.True(t, bytes.Equal(src.MsgData, dst.MsgData))
}

func TestPeerMessageChunks(t *testing.T) {
	var err error
	var src, dst *peering.PeerMessage
	nodeIdentity := ed25519.GenerateKeyPair()
	chunkSize := 100
	chp := chopper.NewChopper()
	data := make([]byte, 2013)
	for i := range data {
		data[i] = byte(rand.Intn(255))
	}
	src = &peering.PeerMessage{
		PeeringID:   peering.RandomPeeringID(),
		SenderIndex: uint16(123),
		Timestamp:   time.Now().UnixNano(),
		MsgType:     peering.FirstUserMsgCode + 17,
		MsgData:     data,
	}
	var chunks [][]byte
	chunks, err = src.ChunkedBytes(chunkSize, chp, &nodeIdentity)
	require.Nil(t, err)
	require.NotNil(t, chunks)
	require.True(t, len(chunks) > 1)
	for i := range chunks {
		var chunkMsg *peering.PeerMessage
		chunkMsg, err = peering.NewPeerMessageFromBytes(chunks[i], &nodeIdentity.PublicKey)
		require.Nil(t, err)
		require.Equal(t, peering.MsgTypeMsgChunk, chunkMsg.MsgType)
		dst, err = peering.NewPeerMessageFromChunks(chunkMsg.MsgData, chunkSize, chp, &nodeIdentity.PublicKey)
		require.Nil(t, err)
		if i == len(chunks)-1 {
			require.NotNil(t, dst)
		} else {
			require.Nil(t, dst)
		}
	}
	require.EqualValues(t, src.PeeringID, dst.PeeringID)
	require.Equal(t, src.SenderIndex, dst.SenderIndex)
	require.Equal(t, src.Timestamp, dst.Timestamp)
	require.Equal(t, src.MsgType, dst.MsgType)
	require.True(t, bytes.Equal(src.MsgData, dst.MsgData))
}
