package peering_test

import (
	"bytes"
	"testing"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/stretchr/testify/require"
)

func TestPeerMessageCodec(t *testing.T) {
	var err error
	var src, dst *peering.PeerMessageNet
	src = &peering.PeerMessageNet{
		PeerMessageData: peering.PeerMessageData{
			PeeringID:   peering.RandomPeeringID(),
			MsgReceiver: byte(10),
			MsgType:     peering.FirstUserMsgCode + 17,
			MsgData:     []byte{1, 2, 3, 4, 5},
		},
	}
	var bin []byte
	bin, err = src.Bytes()
	require.Nil(t, err)
	require.NotNil(t, bin)
	dst, err = peering.NewPeerMessageNetFromBytes(bin)
	require.Nil(t, err)
	require.NotNil(t, dst)
	require.EqualValues(t, src.PeeringID, dst.PeeringID)
	require.Equal(t, src.MsgReceiver, dst.MsgReceiver)
	require.Equal(t, src.MsgType, dst.MsgType)
	require.True(t, bytes.Equal(src.MsgData, dst.MsgData))
}
