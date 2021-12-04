package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"golang.org/x/xerrors"
)

var (
	ErrNoAliasOutputAtIndex0 = xerrors.New("origin AliasOutput not found at index 0")

	nilAliasID iotago.AliasID
)

// GetAnchorFromTransaction analyzes the output at index 0 and extracts anchor information
// Otherwise error
func GetAnchorFromTransaction(tx *iotago.Transaction) (*iscp.StateAnchor, error) {
	anchorOutput, ok := tx.Essence.Outputs[0].(*iotago.AliasOutput)
	if !ok {
		return nil, ErrNoAliasOutputAtIndex0
	}
	txid, err := tx.ID()
	if err != nil {
		return nil, xerrors.Errorf("GetAnchorFromTransaction: %w", err)
	}
	aliasID := anchorOutput.AliasID
	isOrigin := false

	if aliasID == nilAliasID {
		isOrigin = true
		aliasID = iotago.AliasIDFromOutputID(iotago.OutputIDFromTransactionIDAndIndex(*txid, 0))
	}
	sd, err := iscp.StateDataFromBytes(anchorOutput.StateMetadata)
	if err != nil {
		return nil, err
	}
	return &iscp.StateAnchor{
		IsOrigin:             isOrigin,
		OutputID:             iotago.OutputIDFromTransactionIDAndIndex(*txid, 0),
		ChainID:              iscp.ChainIDFromAliasID(aliasID),
		StateController:      anchorOutput.StateController,
		GovernanceController: anchorOutput.GovernanceController,
		StateIndex:           anchorOutput.StateIndex,
		StateData:            sd,
		Deposit:              anchorOutput.Amount,
	}, nil
}
