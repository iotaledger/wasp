package iscmove_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmovetest"
	"github.com/iotaledger/wasp/packages/cryptolib"
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
