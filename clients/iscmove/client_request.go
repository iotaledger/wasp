package iscmove

import (
	"context"
	"fmt"

	"github.com/fardream/go-bcs/bcs"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// CreateAndSendRequest calls <packageID>::request::create_and_send_request() and transfers the created
// Request to the signer.
func (c *Client) CreateAndSendRequest(
	ctx context.Context,
	cryptolibSigner cryptolib.Signer,
	packageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
	assetsBagRef *sui.ObjectRef,
	iscContractName string,
	iscFunctionName string,
	args [][]byte,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	devMode bool,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)
	anchorRes, err := c.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: anchorAddress})
	if err != nil {
		return nil, fmt.Errorf("failed to get anchor ref: %w", err)
	}
	anchorRef := anchorRes.Data.Ref()
	ptb := sui.NewProgrammableTransactionBuilder()

	ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "request",
				Function:      "create_and_send_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(anchorRef.ObjectID),
					ptb.MustObj(sui.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustPure(&bcs.Option[string]{Some: iscContractName}),
					ptb.MustPure(&bcs.Option[string]{Some: iscFunctionName}),
					ptb.MustPure(&bcs.Option[[][]byte]{Some: args}),
				},
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui.NewProgrammable(
		signer.Address(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	return txnResponse, nil
}
