package iotaconn_grpc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
)

func TestEventDecoding(t *testing.T) {
	// Due to the nature of how events work, we would need to start a L1 node supporting gRPC, deploy an ISC contract,
	// emit requests and receive the event. At the time of writing no Docker image is available supporting gRPC.
	// To test event decoding easily, use constant event data and validate the decoding.
	const eventData = "0x15da4c66138800e37bc2e857bee8107091b19a35fec17aeef6a9c6b5aa646d6d936e800c6a4d4a111c465e6850ce3bd0f3741f7280b4c72080ce630d9da4a4e2"

	requestID := lo.Must(iotago.ObjectIDFromHex("0x15da4c66138800e37bc2e857bee8107091b19a35fec17aeef6a9c6b5aa646d6d"))
	anchorID := lo.Must(iotago.AddressFromHex("0x936e800c6a4d4a111c465e6850ce3bd0f3741f7280b4c72080ce630d9da4a4e2"))

	event, err := bcs.Unmarshal[iscmove.RequestEvent](lo.Must(hexutil.Decode(eventData)))
	require.NoError(t, err)

	require.Equal(t, *requestID, event.RequestID)
	require.Equal(t, *anchorID, event.Anchor)
}
