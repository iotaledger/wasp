package peering_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/stretchr/testify/require"
)

func TestPeerMessageCodec(t *testing.T) {
	var err error
	var src, dst *peering.PeerMessage
	src = &peering.PeerMessage{
		PeeringID:   peering.RandomPeeringID(),
		SenderIndex: uint16(123),
		Timestamp:   time.Now().UnixNano(),
		MsgType:     peering.FirstUserMsgCode + 17,
		MsgData:     []byte{1, 2, 3, 4, 5},
	}
	var bin []byte
	bin, err = src.Bytes()
	require.Nil(t, err)
	require.NotNil(t, bin)
	dst, err = peering.NewPeerMessageFromBytes(bin)
	require.Nil(t, err)
	require.NotNil(t, dst)
	require.EqualValues(t, src.PeeringID, dst.PeeringID)
	require.Equal(t, src.SenderIndex, dst.SenderIndex)
	require.Equal(t, src.Timestamp, dst.Timestamp)
	require.Equal(t, src.MsgType, dst.MsgType)
	require.True(t, bytes.Equal(src.MsgData, dst.MsgData))
}
