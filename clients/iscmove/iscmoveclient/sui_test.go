package iscmoveclient_test

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func signAndExecuteTransactionGetObjectRef(
	client *iscmoveclient.Client,
	cryptolibSigner cryptolib.Signer,
	txnBytes []byte,
	module sui.Identifier,
	objName sui.Identifier) (*sui.ObjectRef, error) {
	signer := cryptolib.SignerToSuiSigner(cryptolibSigner)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		signer, txnBytes, &suijsonrpc.SuiTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true})
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	if !txnResponse.Effects.Data.IsSuccess() {
		return nil, fmt.Errorf("failed to execute the transaction: %s", txnResponse.Effects.Data.V1.Status.Error)
	}
	ref, err := txnResponse.GetCreatedObjectInfo(module, objName)
	if err != nil {
		return nil, fmt.Errorf("failed to create AssetsBag: %w", err)
	}
	return ref, nil
}
