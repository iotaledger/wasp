package isctest

import (
	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/samber/lo"
)

var TestChainID = lo.Must(isc.ChainIDFromBytes(testutil.TestBytes(iotago.AddressLen)))
var TestAgentID = isc.NewContractAgentID(isc.Hn("test-contract"))

func testRequestWithRef(ref *iotago.ObjectRef, sender *cryptolib.Address, assetBagID *iotago.Address) *iscmove.RefWithObject[iscmove.Request] {
	return &iscmove.RefWithObject[iscmove.Request]{
		ObjectRef: *ref,
		Object: &iscmove.Request{
			ID:     *ref.ObjectID,
			Sender: sender,
			AssetsBag: iscmove.AssetsBagWithBalances{
				AssetsBag: iscmove.AssetsBag{ID: *assetBagID, Size: 1},
				Assets:    *iscmove.NewAssets(1000),
			},
			Message: iscmove.Message{
				Contract: 123,
				Function: 456,
				Args:     [][]byte{[]byte("testarg1"), []byte("testarg2")},
			},
			AllowanceBCS: bcs.MustMarshal(iscmove.NewAssets(111).
				SetCoin(iotajsonrpc.MustCoinTypeFromString("0x1::coin::TEST_A"), 222)),
			GasBudget: 1000,
		},
	}
}

var TestRequestWithRef = testRequestWithRef(
	iotatest.TestObjectRef,
	cryptolib.TestAddress,
	iotatest.TestAddress,
)
