package serialization

import (
	"context"
	"log"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	suiclient2 "github.com/iotaledger/wasp/clients/iota-go/suiclient"
	suijsonrpc2 "github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	suisigner2 "github.com/iotaledger/wasp/clients/iota-go/suisigner"
)

type Publisher struct {
	client *suiclient2.Client
	signer suisigner2.Signer
}

func NewPublisher(client *suiclient2.Client, signer suisigner2.Signer) *Publisher {
	return &Publisher{
		client: client,
		signer: signer,
	}
}

func (p *Publisher) PublishEvents(ctx context.Context, packageID *sui.PackageID) {
	txnBytes, err := p.client.MoveCall(
		ctx,
		suiclient2.MoveCallRequest{
			Signer:    p.signer.Address(),
			PackageID: packageID,
			Module:    "eventpub",
			Function:  "emit_clock",
			TypeArgs:  []string{},
			Arguments: []any{},
			GasBudget: suijsonrpc2.NewBigInt(100000),
		},
	)
	if err != nil {
		log.Panic(err)
	}

	signature, err := p.signer.SignTransactionBlock(txnBytes.TxBytes.Data(), suisigner2.DefaultIntent())
	if err != nil {
		log.Panic(err)
	}

	txnResponse, err := p.client.ExecuteTransactionBlock(
		ctx, suiclient2.ExecuteTransactionBlockRequest{
			TxDataBytes: txnBytes.TxBytes.Data(),
			Signatures:  []*suisigner2.Signature{signature},
			Options: &suijsonrpc2.SuiTransactionBlockResponseOptions{
				ShowInput:          true,
				ShowEffects:        true,
				ShowEvents:         true,
				ShowObjectChanges:  true,
				ShowBalanceChanges: true,
			},
			RequestType: suijsonrpc2.TxnRequestTypeWaitForLocalExecution,
		},
	)
	if err != nil {
		log.Panic(err)
	}

	log.Println(txnResponse)
}
