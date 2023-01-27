package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func nativeTokenOutputMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, keyNativeTokenOutputMap)
}

func nativeTokenOutputMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, keyNativeTokenOutputMap)
}

// SaveNativeTokenOutput map nativeTokenID -> foundryRec
func SaveNativeTokenOutput(state kv.KVStore, out *iotago.BasicOutput, blockIndex uint32, outputIndex uint16) {
	tokenRec := nativeTokenOutputRec{
		StorageBaseTokens: out.Amount,
		Amount:            out.NativeTokens[0].Amount,
		BlockIndex:        blockIndex,
		OutputIndex:       outputIndex,
	}
	nativeTokenOutputMap(state).MustSetAt(out.NativeTokens[0].ID[:], tokenRec.Bytes())
}

func DeleteNativeTokenOutput(state kv.KVStore, nativeTokenID iotago.NativeTokenID) {
	nativeTokenOutputMap(state).MustDelAt(nativeTokenID[:])
}

func GetNativeTokenOutput(state kv.KVStoreReader, nativeTokenID iotago.NativeTokenID, chainID isc.ChainID) (*iotago.BasicOutput, uint32, uint16) {
	data := nativeTokenOutputMapR(state).MustGetAt(nativeTokenID[:])
	if data == nil {
		return nil, 0, 0
	}
	tokenRec := mustNativeTokenOutputRecFromBytes(data)
	ret := &iotago.BasicOutput{
		Amount: tokenRec.StorageBaseTokens,
		NativeTokens: iotago.NativeTokens{{
			ID:     nativeTokenID,
			Amount: tokenRec.Amount,
		}},
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: chainID.AsAddress()},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				Address: chainID.AsAddress(),
			},
		},
	}
	return ret, tokenRec.BlockIndex, tokenRec.OutputIndex
}
