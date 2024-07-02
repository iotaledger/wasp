package serialization

import (
	"context"
	"log"

	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type Publisher struct {
	client *suiclient.Client
	signer suisigner.Signer
}

func NewPublisher(client *suiclient.Client, signer suisigner.Signer) *Publisher {
	return &Publisher{
		client: client,
		signer: signer,
	}
}

func (p *Publisher) PublishEvents(ctx context.Context, packageID *sui.PackageID) {
	txnBytes, err := p.client.MoveCall(
		ctx,
		suiclient.MoveCallRequest{
			Signer:    p.signer.Address(),
			PackageID: packageID,
			Module:    "eventpub",
			Function:  "emit_clock",
			TypeArgs:  []string{},
			Arguments: []any{},
			GasBudget: suijsonrpc.NewBigInt(100000),
		},
	)
	if err != nil {
		log.Panic(err)
	}

	signature, err := p.signer.SignTransactionBlock(txnBytes.TxBytes.Data(), suisigner.DefaultIntent())
	if err != nil {
		log.Panic(err)
	}

	txnResponse, err := p.client.ExecuteTransactionBlock(
		ctx, suiclient.ExecuteTransactionBlockRequest{
			TxDataBytes: txnBytes.TxBytes.Data(),
			Signatures:  []*suisigner.Signature{signature},
			Options: &suijsonrpc.SuiTransactionBlockResponseOptions{
				ShowInput:          true,
				ShowEffects:        true,
				ShowEvents:         true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
			RequestType: suijsonrpc.TxnRequestTypeWaitForLocalExecution,
		},
	)
	if err != nil {
		log.Panic(err)
	}

	log.Println(txnResponse)
}
