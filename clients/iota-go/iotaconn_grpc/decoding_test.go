package iotaconn_grpc

import (
	"encoding/json"
	"fmt"
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
	const eventData = "0x209aa3936ae9669d7c6ff38b40804f0f600f128c164d71f55029e5f005acff8cb30000000000000000a8d1cc681912e49f23511e8d113f329c711bbe27d4d43eb80e109260a4c1375a0772657175657374290ebea1b6096c9e624a5826d4ae7c455da95219ac2658c2c7dd091d5584e9afa8d1cc681912e49f23511e8d113f329c711bbe27d4d43eb80e109260a4c1375a07726571756573740c526571756573744576656e7400a1017b22616e63686f72223a22307832316564663232663734326665393864346661313365336664653166396639643466363832633835613833646338396530643031663962303361613437373963222c22726571756573745f6964223a22307837653737333536353930386463653532653736393639623563626463313737656461633565313233333133663839623662363537366337376334633935336134227d407e773565908dce52e76969b5cbdc177edac5e123313f89b6b6576c77c4c953a421edf22f742fe98d4fa13e3fde1f9f9d4f682c85a83dc89e0d01f9b03aa4779c01c19a757498010000"

	iscPackageID := lo.Must(
		iotago.PackageIDFromHex(
			"0xa8d1cc681912e49f23511e8d113f329c711bbe27d4d43eb80e109260a4c1375a",
		),
	)

	senderAddress := lo.Must(
		iotago.AddressFromHex(
			"0x290ebea1b6096c9e624a5826d4ae7c455da95219ac2658c2c7dd091d5584e9af",
		),
	)

	requestID := lo.Must(iotago.ObjectIDFromHex("0x7e773565908dce52e76969b5cbdc177edac5e123313f89b6b6576c77c4c953a4"))
	anchorID := lo.Must(iotago.AddressFromHex("0x21edf22f742fe98d4fa13e3fde1f9f9d4f682c85a83dc89e0d01f9b03aa4779c"))

	event, err := bcs.Unmarshal[IotaRpcEvent](lo.Must(hexutil.Decode(eventData)))
	require.NoError(t, err)

	fmt.Printf(
		"Event Data:\nEvent Digest: %v\n Event Sequence ID: %v\n PackageID: %v\n Transaction Module: %v\n Sender: %v\n"+
			" Type:\n  Name: %v\n  TypeModule: %v\n  TypeParams: %v\n ParsedJson: %v\n BCS: %v\n Timestamp: %v\n\n",
		event.EventID.Digest.String(),
		event.EventID.EventSequence,
		event.PackageID.String(),
		event.TransactionModule,
		event.Sender.String(),
		event.Type.Name, event.Type.Module, event.Type.Params,
		string(event.ParsedJson),
		hexutil.Encode(event.Bcs),
		*event.Timestamp,
	)

	require.Equal(t, uint64(0), event.EventID.EventSequence)
	require.Equal(t, "BQeTG78q9YomHxp3gn1dNCpLoN7tPNFedqtAUuqva1HU", event.EventID.Digest.String())

	require.Equal(t, *iscPackageID, event.PackageID)
	require.Equal(t, *senderAddress, event.Sender)
	require.Equal(t, iscmove.RequestEventObjectName, event.Type.Name)
	require.Equal(t, iscmove.RequestModuleName, event.Type.Module)

	reqEventBCS, err := bcs.Unmarshal[iscmove.RequestEvent](event.Bcs)
	require.NoError(t, err)

	require.Equal(t, *requestID, reqEventBCS.RequestID)
	require.Equal(t, *anchorID, reqEventBCS.Anchor)

	var reqEventJSON iscmove.RequestEvent
	err = json.Unmarshal(event.ParsedJson, &reqEventJSON)
	require.NoError(t, err)

	require.Equal(t, *requestID, reqEventJSON.RequestID)
	require.Equal(t, *anchorID, reqEventJSON.Anchor)
}
