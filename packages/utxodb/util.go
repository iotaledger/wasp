package utxodb

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"golang.org/x/xerrors"
)

func GetSingleChainedAliasOutput(tx *iotago.Transaction) (*iotago.AliasOutput, iotago.OutputID, error) {
	outputs, err := tx.OutputsSet()
	if err != nil {
		return nil, iotago.OutputID{}, err
	}
	var rOutput *iotago.AliasOutput
	var rID iotago.OutputID
	var count int
	for id, output := range outputs {
		var ok bool
		rOutput, ok = output.(*iotago.AliasOutput)
		if ok {
			rID = id
			count++
		}
	}
	if count == 0 {
		return nil, iotago.OutputID{}, nil
	} else if count > 1 {
		return nil, iotago.OutputID{}, xerrors.New("more than one chained output was found")
	}
	return rOutput, rID, nil
}
