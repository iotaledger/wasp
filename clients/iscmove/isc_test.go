package iscmove_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

func TestIscCodec(t *testing.T) {
	type ExampleObj struct {
		A int
	}

	bcs.TestCodec(t, iscmove.RefWithObject[ExampleObj]{
		ObjectRef: *iotatest.RandomObjectRef(),
		Object:    &ExampleObj{A: 42},
		Owner:     iotago.MustAddressFromHex(testcommon.TestAddress),
	})

	bcs.TestCodecAndHash(t, iscmove.RefWithObject[ExampleObj]{
		ObjectRef: *iotatest.TestObjectRef,
		Object:    &ExampleObj{A: 42},
		Owner:     iotago.MustAddressFromHex(testcommon.TestAddress),
	}, "15ca3116a5d0")

	anchor := iscmovetest.RandomAnchor()

	anchorRef := iscmove.RefWithObject[iscmove.Anchor]{
		ObjectRef: iotago.ObjectRef{
			ObjectID: &anchor.ID,
			Version:  13,
			Digest:   iotatest.RandomDigest(),
		},
		Object: &anchor,
		Owner:  iotago.MustAddressFromHex(testcommon.TestAddress),
	}

	bcs.TestCodec(t, anchorRef)

	anchorRef.Object = &iscmovetest.TestAnchor
	anchorRef.ObjectID = &anchorRef.Object.ID
	anchorRef.Digest = iotatest.TestDigest

	// Changelog
	// <lmoe> 20.08.25 changed 2ed70074c011 to 2750607f6325 to adjust DefaultGasFeePolicy / MinGasPerRequest
	bcs.TestCodecAndHash(t, anchorRef, "2750607f6325")

	bcs.TestCodecAndHash(t, iscmove.AssetsBagWithBalances{
		AssetsBag: iscmovetest.TestAssetsBag,
		Assets: *iscmove.NewAssets(123456).
			SetCoin(iotajsonrpc.MustCoinTypeFromString("0x1::a::A"), 100).
			AddObject(*iotatest.TestAddress, iotago.MustTypeFromString("0x2::a::B")),
	}, "17fd55be42d7")
}

func TestUnmarshalBCS(t *testing.T) {
	req := iscmoveclient.MoveRequest{
		ID:     *iotatest.RandomAddress(),
		Sender: cryptolib.NewAddressFromIota(iotatest.RandomAddress()),
		AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
			ID: *iotatest.RandomAddress(),
			Value: &iscmove.AssetsBagWithBalances{
				AssetsBag: iscmovetest.RandomAssetsBag(),
				Assets: *iscmove.NewAssets(0).
					AddObject(iscmovetest.RandomAnchor().ID, iotago.MustTypeFromString("0x1::a::A")),
			},
		},
		Message: *iscmovetest.RandomMessage(),
		Allowance: bcs.MustMarshal(iscmove.NewAssets(0).
			SetCoin(iotajsonrpc.IotaCoinType, 100).
			AddObject(iotago.ObjectID{}, iotago.MustTypeFromString("0x1::a::A"))),
		GasBudget: 100,
	}
	b, err := bcs.Marshal(&req)
	require.NoError(t, err)

	var targetReq iscmoveclient.MoveRequest

	err = iotaclient.UnmarshalBCS(b, &targetReq)
	require.Nil(t, err)
}
