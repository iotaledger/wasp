package isc_test

import (
	"testing"
)

func TestOnLedgerCodec(t *testing.T) {
	// TODO: Assets and AssetBag were not in Read/Write functions, but they are set in constructor.
	// And they are unexpeorted. Thus reading will be always asymmetric to writing.
	// SO we can't test it.

	// onledgerRequest, err := isc.OnLedgerFromRequest(&iscmove.RefWithObject[iscmove.Request]{
	// 	ObjectRef: *sui.RandomObjectRef(),
	// 	Object: &iscmove.Request{
	// 		ID:     *sui.RandomAddress(),
	// 		Sender: cryptolib.NewRandomAddress(),
	// 		Message: iscmove.Message{
	// 			Contract: 123,
	// 			Function: 456,
	// 			Args: [][]byte{
	// 				{1, 2, 3},
	// 				{4, 5, 6},
	// 			},
	// 		},
	// 		// AssetsBag: iscmove.Referent[iscmove.AssetsBagWithBalances]{
	// 		// 	ID: *sui.RandomAddress(),
	// 		// 	Value: &iscmove.AssetsBagWithBalances{
	// 		// 		AssetsBag: iscmove.AssetsBag{
	// 		// 			ID:   *sui.RandomAddress(),
	// 		// 			Size: 567,
	// 		// 		},
	// 		// 	},
	// 		// },
	// 		GasBudget: 890,
	// 		Allowance: iscmove.Referent[iscmove.Allowance]{
	// 			ID: *sui.RandomAddress(),
	// 			Value: &iscmove.Allowance{
	// 				ID: *sui.RandomAddress(),
	// 			},
	// 		},
	// 	},
	// }, cryptolib.NewRandomAddress())
	// require.NoError(t, err)

	// bcs.TestCodec(t, onledgerRequest)
}
