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
	var targetReq iscmoveclient.MoveRequest
	req := iscmoveclient.MoveRequest{
		ID:     *iotatest.RandomAddress(),
		Sender: cryptolib.NewAddressFromIota(iotatest.RandomAddress()),
		AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
			ID: *iotatest.RandomAddress(),
			Value: &iscmove.AssetsBagWithBalances{
				AssetsBag: iscmovetest.RandomAssetsBag(),
				Balances:  iscmove.AssetsBagBalances{},
			},
		},
		Message: *iscmovetest.RandomMessage(),
		Allowance: []iscmove.CoinAllowance{
			{CoinType: iotajsonrpc.IotaCoinType, Balance: 100},
			{CoinType: "0x1:AB:ab", Balance: 200},
		},
		GasBudget: 100,
	}
	b, err := bcs.Marshal(&req)
	require.NoError(t, err)
	err = iotaclient.UnmarshalBCS(b, &targetReq)
	require.NotNil(t, err)
}
