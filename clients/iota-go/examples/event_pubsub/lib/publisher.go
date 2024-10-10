package serialization

import (
	"context"
	"log"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
)

type Publisher struct {
	client *iotaclient.Client
	signer iotasigner.Signer
}

func NewPublisher(client *iotaclient.Client, signer iotasigner.Signer) *Publisher {
	return &Publisher{
		client: client,
		signer: signer,
	}
}

func (p *Publisher) PublishEvents(ctx context.Context, packageID *iotago.PackageID) {
	txnBytes, err := p.client.MoveCall(
		ctx,
		iotaclient.MoveCallRequest{
			Signer:    p.signer.Address(),
			PackageID: packageID,
			Module:    "eventpub",
			Function:  "emit_clock",
			TypeArgs:  []string{},
			Arguments: []any{},
			GasBudget: iotajsonrpc.NewBigInt(100000),
		},
	)
	if err != nil {
		log.Panic(err)
	}

	signature, err := p.signer.SignTransactionBlock(txnBytes.TxBytes.Data(), iotasigner.DefaultIntent())
	if err != nil {
		log.Panic(err)
	}

	txnResponse, err := p.client.ExecuteTransactionBlock(
		ctx, iotaclient.ExecuteTransactionBlockRequest{
			TxDataBytes: txnBytes.TxBytes.Data(),
			Signatures:  []*iotasigner.Signature{signature},
			Options: &iotajsonrpc.SuiTransactionBlockResponseOptions{
				ShowInput:          true,
				ShowEffects:        true,
				ShowEvents:         true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		},
	)
	if err != nil {
		log.Panic(err)
	}

	log.Println(txnResponse)
}
