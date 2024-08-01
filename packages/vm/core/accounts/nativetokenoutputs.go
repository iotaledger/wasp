package accounts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/sui-go/sui"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func (s *StateWriter) newNativeTokensArray() *collections.Array {
	return collections.NewArray(s.state, keyNewNativeTokens)
}

func (s *StateWriter) nativeTokenOutputMap() *collections.Map {
	return collections.NewMap(s.state, keyNativeTokenOutputMap)
}

func (s *StateReader) nativeTokenOutputMapR() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, keyNativeTokenOutputMap)
}

// SaveNativeTokenOutput map nativeTokenID -> foundryRec
func (s *StateWriter) SaveNativeTokenOutput(out *iotago.BasicOutput, outputIndex uint16) {
	tokenRec := nativeTokenOutputRec{
		// TransactionID is unknown yet, will be filled next block
		OutputID:          sui.ObjectID{},
		StorageBaseTokens: out.Amount,
		Amount:            out.NativeTokens[0].Amount,
	}
	s.nativeTokenOutputMap().SetAt(out.NativeTokens[0].ID[:], tokenRec.Bytes())
	s.newNativeTokensArray().Push(out.NativeTokens[0].ID[:])
}

func (s *StateWriter) updateNativeTokenOutputIDs(anchorTxID sui.ObjectID) {
	newNativeTokens := s.newNativeTokensArray()
	allNativeTokens := s.nativeTokenOutputMap()
	n := newNativeTokens.Len()
	for i := uint32(0); i < n; i++ {
		k := newNativeTokens.GetAt(i)
		rec := mustNativeTokenOutputRecFromBytes(allNativeTokens.GetAt(k))
		rec.OutputID = anchorTxID
		allNativeTokens.SetAt(k, rec.Bytes())
	}
	newNativeTokens.Erase()
}

func (s *StateWriter) DeleteNativeTokenOutput(nativeTokenID isc.CoinType) {
	s.nativeTokenOutputMap().DelAt(nativeTokenID.Bytes())
}

func (s *StateReader) GetNativeTokenOutput(nativeTokenID sui.ObjectID, chainID isc.ChainID) (*iotago.BasicOutput, iotago.OutputID) {
	data := s.nativeTokenOutputMapR().GetAt(nativeTokenID[:])
	if data == nil {
		return nil, iotago.OutputID{}
	}
	tokenRec := mustNativeTokenOutputRecFromBytes(data)

	panic("refactor me: AsIotagoAddress")

	ret := &iotago.BasicOutput{
		Amount: tokenRec.StorageBaseTokens,
		NativeTokens: isc.NativeTokens{{
			ID:     nativeTokenID,
			Amount: tokenRec.Amount,
		}},
		Conditions: iotago.UnlockConditions{
			//&iotago.AddressUnlockCondition{Address: chainID.AsAddress().AsIotagoAddress()},
		},
		Features: iotago.Features{
			&iotago.SenderFeature{
				//	Address: chainID.AsAddress().AsIotagoAddress(),
			},
		},
	}
	return ret, tokenRec.OutputID
}
