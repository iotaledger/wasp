package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

func GetAnchorFromTransaction(tx *iotago.Transaction, chainID *iscp.ChainID) (*iscp.StateAnchor, error) {
	var anchorOutput *iotago.AliasOutput
	var idx uint16

	if chainID == nil {
		return getOriginAnchor(tx)
	}
	aliasID := chainID.AsAliasID()
	for i, out := range tx.Essence.Outputs {
		if a, ok := out.(*iotago.AliasOutput); ok {
			if a.AliasID == *aliasID {
				anchorOutput = a
				idx = uint16(i)
				break
			}
		}
	}
	if anchorOutput == nil {
		ret, err := getOriginAnchor(tx)
		if err != nil {
			return nil, xerrors.Errorf("alias id wasn't found in transaction: %s", chainID.String())
		}
		return ret, nil
	}
	sd, err := iscp.StateDataFromBytes(anchorOutput.StateMetadata)
	if err != nil {
		return nil, err
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, err
	}
	return &iscp.StateAnchor{
		IsOrigin:             false,
		StateController:      anchorOutput.StateController,
		GovernanceController: anchorOutput.GovernanceController,
		StateIndex:           anchorOutput.StateIndex,
		OutputID: iotago.UTXOInput{
			TransactionID:          *txid,
			TransactionOutputIndex: idx,
		},
		StateData: sd,
		Deposit:   anchorOutput.Amount,
	}, nil
}

func getOriginAnchor(tx *iotago.Transaction) (*iscp.StateAnchor, error) {
	txid, err := tx.ID()
	if err != nil {
		return nil, err
	}
	var anchorOutput *iotago.AliasOutput
	var nilAliasID iotago.AliasID
	var idx uint16
	for i, out := range tx.Essence.Outputs {
		if a, ok := out.(*iotago.AliasOutput); ok {
			if a.AliasID == nilAliasID {
				idx = uint16(i)
				anchorOutput = a
				break
			}
		}
	}
	if anchorOutput == nil {
		return nil, xerrors.New("origin AliasOutput not found")
	}
	sd, err := iscp.StateDataFromBytes(anchorOutput.StateMetadata)
	if err != nil {
		return nil, err
	}
	return &iscp.StateAnchor{
		IsOrigin:             true,
		StateController:      anchorOutput.StateController,
		GovernanceController: anchorOutput.GovernanceController,
		StateIndex:           anchorOutput.StateIndex,
		OutputID: iotago.UTXOInput{
			TransactionID:          *txid,
			TransactionOutputIndex: idx,
		},
		StateData: sd,
		Deposit:   anchorOutput.Amount,
	}, nil
}
