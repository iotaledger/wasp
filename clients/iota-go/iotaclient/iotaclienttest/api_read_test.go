package iotaclienttest

import (
	"context"
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetChainIdentifier(t *testing.T) {
	client := l1starter.Instance().L1Client()
	_, err := client.GetChainIdentifier(context.Background())
	require.NoError(t, err)
}

func TestGetCheckpoint(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sn := iotajsonrpc.NewBigInt(3)
	checkpoint, err := client.GetCheckpoint(context.Background(), sn)
	require.NoError(t, err)
	// targetCheckpoint := &iotajsonrpc.Checkpoint{
	// 	Epoch:                    iotajsonrpc.NewBigInt(0),
	// 	SequenceNumber:           iotajsonrpc.NewBigInt(1000),
	// 	Digest:                   *iotago.MustNewDigest("Eu7yhUZ1oma3fk8KhHW86usFvSmjZ7QPEhPsX7ZYfRg3"),
	// 	NetworkTotalTransactions: iotajsonrpc.NewBigInt(1004),
	// 	PreviousDigest:           iotago.MustNewDigest("AcrgtLsNQxZQRU1JK395vanZzSR6nTun6huJAxEJuk14"),
	// 	EpochRollingGasCostSummary: iotajsonrpc.GasCostSummary{
	// 		ComputationCost:         iotajsonrpc.NewBigInt(0),
	// 		StorageCost:             iotajsonrpc.NewBigInt(0),
	// 		StorageRebate:           iotajsonrpc.NewBigInt(0),
	// 		NonRefundableStorageFee: iotajsonrpc.NewBigInt(0),
	// 	},
	// 	TimestampMs:           iotajsonrpc.NewBigInt(1725548499477),
	// 	Transactions:          []*iotago.Digest{iotago.MustNewDigest("8iu72fMHEFHiJMfjrPDTKBPufQgMSRKfeh2idG5CoHvE")},
	// 	CheckpointCommitments: []iotago.CheckpointCommitment{},
	// 	ValidatorSignature:    *iotago.MustNewBase64Data("k0u7tZR87vS8glhPgmCzgKFm1UU1ikmPmO9nVzFXn9XY20kpftc6zxdBe0lmSAzs"),
	// }

	require.Equal(t, sn, checkpoint.SequenceNumber)
}

func TestGetCheckpoints(t *testing.T) {
	client := l1starter.Instance().L1Client()
	cursor := iotajsonrpc.NewBigInt(999)
	limit := uint64(2)
	checkpointPage, err := client.GetCheckpoints(
		context.Background(), iotaclient.GetCheckpointsRequest{
			Cursor: cursor,
			Limit:  &limit,
		},
	)
	require.NoError(t, err)
	// targetCheckpoints := []*iotajsonrpc.Checkpoint{
	// 	{
	// 		Epoch:                    iotajsonrpc.NewBigInt(0),
	// 		SequenceNumber:           iotajsonrpc.NewBigInt(1000),
	// 		Digest:                   *iotago.MustNewDigest("Eu7yhUZ1oma3fk8KhHW86usFvSmjZ7QPEhPsX7ZYfRg3"),
	// 		NetworkTotalTransactions: iotajsonrpc.NewBigInt(1004),
	// 		PreviousDigest:           iotago.MustNewDigest("AcrgtLsNQxZQRU1JK395vanZzSR6nTun6huJAxEJuk14"),
	// 		EpochRollingGasCostSummary: iotajsonrpc.GasCostSummary{
	// 			ComputationCost:         iotajsonrpc.NewBigInt(0),
	// 			StorageCost:             iotajsonrpc.NewBigInt(0),
	// 			StorageRebate:           iotajsonrpc.NewBigInt(0),
	// 			NonRefundableStorageFee: iotajsonrpc.NewBigInt(0),
	// 		},
	// 		TimestampMs:           iotajsonrpc.NewBigInt(1725548499477),
	// 		Transactions:          []*iotago.Digest{iotago.MustNewDigest("8iu72fMHEFHiJMfjrPDTKBPufQgMSRKfeh2idG5CoHvE")},
	// 		CheckpointCommitments: []iotago.CheckpointCommitment{},
	// 		ValidatorSignature:    *iotago.MustNewBase64Data("k0u7tZR87vS8glhPgmCzgKFm1UU1ikmPmO9nVzFXn9XY20kpftc6zxdBe0lmSAzs"),
	// 	},
	// 	{
	// 		Epoch:                    iotajsonrpc.NewBigInt(0),
	// 		SequenceNumber:           iotajsonrpc.NewBigInt(1001),
	// 		Digest:                   *iotago.MustNewDigest("EJtUUwsKXJR9C9JcJ31e3VZ5rPEsjRu4cSMUaGiTARyo"),
	// 		NetworkTotalTransactions: iotajsonrpc.NewBigInt(1005),
	// 		PreviousDigest:           iotago.MustNewDigest("Eu7yhUZ1oma3fk8KhHW86usFvSmjZ7QPEhPsX7ZYfRg3"),
	// 		EpochRollingGasCostSummary: iotajsonrpc.GasCostSummary{
	// 			ComputationCost:         iotajsonrpc.NewBigInt(0),
	// 			StorageCost:             iotajsonrpc.NewBigInt(0),
	// 			StorageRebate:           iotajsonrpc.NewBigInt(0),
	// 			NonRefundableStorageFee: iotajsonrpc.NewBigInt(0),
	// 		},
	// 		TimestampMs:           iotajsonrpc.NewBigInt(1725548500033),
	// 		Transactions:          []*iotago.Digest{iotago.MustNewDigest("X3QFYvZm5yAgg3nPVPox6jWskpd2cw57Xg8uXNtCTW5")},
	// 		CheckpointCommitments: []iotago.CheckpointCommitment{},
	// 		ValidatorSignature:    *iotago.MustNewBase64Data("jHdu/+su0PZ+93y7du1LH48p1+WAqVm2+5EpvMaFrRBnT0Y63EOTl6fMJFwHEizu"),
	// 	},
	// }
	t.Log(checkpointPage)
}

func TestGetEvents(t *testing.T) {
	t.Skip("TODO: refactor when we have some events")

	client := l1starter.Instance().L1Client()
	digest, err := iotago.NewDigest("3vVi8XZgNpzQ34PFgwJTQqWtPMU84njcBX1EUxUHhyDk")
	require.NoError(t, err)
	events, err := client.GetEvents(context.Background(), digest)
	require.NoError(t, err)
	require.Len(t, events, 1)
	for _, event := range events {
		require.Equal(t, digest, &event.Id.TxDigest)
		require.Equal(
			t,
			iotago.MustPackageIDFromHex("0x000000000000000000000000000000000000000000000000000000000000dee9"),
			event.PackageId,
		)
		require.Equal(t, "clob_v2", event.TransactionModule)
		require.Equal(
			t,
			iotago.MustAddressFromHex("0xf0f13f7ef773c6246e87a8f059a684d60773f85e992e128b8272245c38c94076"),
			event.Sender,
		)
		targetStructTag := iotago.StructTag{
			Address: iotago.MustAddressFromHex("0xdee9"),
			Module:  iotago.Identifier("clob_v2"),
			Name:    iotago.Identifier("OrderPlaced"),
			TypeParams: []iotago.TypeTag{
				{
					Struct: &iotago.StructTag{
						Address: iotago.MustAddressFromHex("0x2"),
						Module:  iotago.Identifier("iota"),
						Name:    iotago.Identifier("IOTA"),
					},
				},
				{
					Struct: &iotago.StructTag{
						Address: iotago.MustAddressFromHex("0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf"),
						Module:  iotago.Identifier("coin"),
						Name:    iotago.Identifier("COIN"),
					},
				},
			},
		}
		require.Equal(t, targetStructTag.Address, event.Type.Address)
		require.Equal(t, targetStructTag.Module, event.Type.Module)
		require.Equal(t, targetStructTag.Name, event.Type.Name)
		require.Equal(t, targetStructTag.TypeParams[0].Struct.Address, event.Type.TypeParams[0].Struct.Address)
		require.Equal(t, targetStructTag.TypeParams[0].Struct.Module, event.Type.TypeParams[0].Struct.Module)
		require.Equal(t, targetStructTag.TypeParams[0].Struct.Name, event.Type.TypeParams[0].Struct.Name)
		require.Equal(t, targetStructTag.TypeParams[0].Struct.TypeParams, event.Type.TypeParams[0].Struct.TypeParams)
		require.Equal(t, targetStructTag.TypeParams[1].Struct.Address, event.Type.TypeParams[1].Struct.Address)
		require.Equal(t, targetStructTag.TypeParams[1].Struct.Module, event.Type.TypeParams[1].Struct.Module)
		require.Equal(t, targetStructTag.TypeParams[1].Struct.Name, event.Type.TypeParams[1].Struct.Name)
		require.Equal(t, targetStructTag.TypeParams[1].Struct.TypeParams, event.Type.TypeParams[1].Struct.TypeParams)
		targetBcsBase64, err := base64.StdEncoding.DecodeString(
			"RAW1DXkf0zRnVOgXGqq2vC7SbCxG790DPBSzCuUHrDObF2oAAAAAgDaEAkYyhy8PAPR7xPX" +
				"+lV7LBzuZSXWnlDlx1Jfi/kERQQnEXcSfTZAuAHT+QdwAAAAAdP5B3AAAALycEAAAAAAAqXmyiI8BAAA=",
		)
		require.NoError(t, err)
		require.Equal(t, targetBcsBase64, event.Bcs.Data())
	}
}

func TestGetLatestCheckpointSequenceNumber(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sequenceNumber, err := client.GetLatestCheckpointSequenceNumber(context.Background())
	require.NoError(t, err)
	num, err := strconv.Atoi(sequenceNumber)
	require.NoError(t, err)
	require.Greater(t, num, 0)
}

func TestGetObject(t *testing.T) {
	type args struct {
		ctx   context.Context
		objID *iotago.ObjectID
	}
	api := l1starter.Instance().L1Client()
	coins, err := api.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: iotago.MustAddressFromHex(testcommon.TestAddress),
			Limit: 1,
		},
	)
	require.NoError(t, err)

	tests := []struct {
		name    string
		api     clients.L1Client
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test for devnet",
			api:  api,
			args: args{
				ctx:   context.Background(),
				objID: coins.Data[0].CoinObjectID,
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.api.GetObject(
					tt.args.ctx, iotaclient.GetObjectRequest{
						ObjectID: tt.args.objID,
						Options: &iotajsonrpc.IotaObjectDataOptions{
							ShowType:                true,
							ShowOwner:               true,
							ShowContent:             true,
							ShowDisplay:             true,
							ShowBcs:                 true,
							ShowPreviousTransaction: true,
							ShowStorageRebate:       true,
						},
					},
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetObject() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				t.Logf("%+v", got)
			},
		)
	}
}

func TestGetProtocolConfig(t *testing.T) {
	api := l1starter.Instance().L1Client()
	version := iotajsonrpc.NewBigInt(1)
	protocolConfig, err := api.GetProtocolConfig(context.Background(), version)
	require.NoError(t, err)
	require.Equal(t, uint64(1), protocolConfig.ProtocolVersion.Uint64())
}

func TestGetTotalTransactionBlocks(t *testing.T) {
	api := l1starter.Instance().L1Client()
	res, err := api.GetTotalTransactionBlocks(context.Background())
	require.NoError(t, err)
	t.Log(res)
}

func TestGetTransactionBlock(t *testing.T) {
	t.Skip("TODO: fix it when the chain is stable. Currently addresses are not stable")
	client := l1starter.Instance().L1Client()
	digest, err := iotago.NewDigest("FGpDhznVR2RpUZG7qB5ZEtME3dH3VL81rz2wFRCuoAv9")
	require.NoError(t, err)
	resp, err := client.GetTransactionBlock(
		context.Background(), iotaclient.GetTransactionBlockRequest{
			Digest: digest,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowInput:          true,
				ShowRawInput:       true,
				ShowEffects:        true,
				ShowRawEffects:     true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
				ShowEvents:         true,
			},
		},
	)
	require.NoError(t, err)

	require.True(t, resp.Effects.Data.IsSuccess())
	require.Greater(t, resp.Effects.Data.V1.ExecutedEpoch.Int64(), 0)
}

func TestMultiGetObjects(t *testing.T) {
	api := l1starter.Instance().L1Client()
	coins, err := api.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: iotago.MustAddressFromHex(testcommon.TestAddress),
			Limit: 1,
		},
	)
	require.NoError(t, err)
	if len(coins.Data) == 0 {
		t.Log("Warning: No Object Id for test.")
		return
	}

	obj := coins.Data[0].CoinObjectID
	objs := []*iotago.ObjectID{obj, obj}
	resp, err := api.MultiGetObjects(
		context.Background(), iotaclient.MultiGetObjectsRequest{
			ObjectIDs: objs,
			Options: &iotajsonrpc.IotaObjectDataOptions{
				ShowType:                true,
				ShowOwner:               true,
				ShowContent:             true,
				ShowDisplay:             true,
				ShowBcs:                 true,
				ShowPreviousTransaction: true,
				ShowStorageRebate:       true,
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, len(objs), len(resp))
	require.Equal(t, resp[0], resp[1])
}

func TestMultiGetTransactionBlocks(t *testing.T) {
	client := l1starter.Instance().L1Client()

	resp, err := client.MultiGetTransactionBlocks(
		context.Background(),
		iotaclient.MultiGetTransactionBlocksRequest{
			Digests: []*iotago.Digest{
				iotago.MustNewDigest("6A3ckipsEtBSEC5C53AipggQioWzVDbs9NE1SPvqrkJr"),
				iotago.MustNewDigest("8AL88Qgk7p6ny3MkjzQboTvQg9SEoWZq4rknEPeXQdH5"),
			},
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects: true,
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, resp, 2)
	require.Equal(t, "6A3ckipsEtBSEC5C53AipggQioWzVDbs9NE1SPvqrkJr", resp[0].Digest.String())
	require.Equal(t, "8AL88Qgk7p6ny3MkjzQboTvQg9SEoWZq4rknEPeXQdH5", resp[1].Digest.String())
}

func TestTryGetPastObject(t *testing.T) {
	// This test might work in general, but can not be executed on either the L1 starter,
	// nor on Alphanet as objects can vanish at any time
	t.Skip()

	api := l1starter.Instance().L1Client()
	// there is no software-level guarantee/SLA that objects with past versions can be retrieved by this API
	resp, err := api.TryGetPastObject(
		context.Background(), iotaclient.TryGetPastObjectRequest{
			ObjectID: iotago.MustObjectIDFromHex("0xdaa46292632c3c4d8f31f23ea0f9b36a28ff3677e9684980e4438403a67a3d8f"),
			Version:  187584506,
			Options: &iotajsonrpc.IotaObjectDataOptions{
				ShowType:  true,
				ShowOwner: true,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp.Data.ObjectNotExists)
}

func TestTryMultiGetPastObjects(t *testing.T) {
	// This test might work in general, but can not be executed on either the L1 starter,
	// nor on Alphanet as objects can vanish at any time
	t.Skip()

	api := l1starter.Instance().L1Client()
	req := []*iotajsonrpc.IotaGetPastObjectRequest{
		{
			ObjectId: iotago.MustObjectIDFromHex("0xdaa46292632c3c4d8f31f23ea0f9b36a28ff3677e9684980e4438403a67a3d8f"),
			Version:  iotajsonrpc.NewBigInt(187584506),
		},
		{
			ObjectId: iotago.MustObjectIDFromHex("0xdaa46292632c3c4d8f31f23ea0f9b36a28ff3677e9684980e4438403a67a3d8f"),
			Version:  iotajsonrpc.NewBigInt(187584500),
		},
	}
	// there is no software-level guarantee/SLA that objects with past versions can be retrieved by this API
	resp, err := api.TryMultiGetPastObjects(
		context.Background(), iotaclient.TryMultiGetPastObjectsRequest{
			PastObjects: req,
			Options: &iotajsonrpc.IotaObjectDataOptions{
				ShowType:  true,
				ShowOwner: true,
			},
		},
	)
	require.NoError(t, err)
	for _, data := range resp {
		require.NotNil(t, data.Data.ObjectNotExists)
	}
}
