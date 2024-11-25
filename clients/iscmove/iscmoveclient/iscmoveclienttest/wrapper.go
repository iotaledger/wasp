package iscmoveclienttest

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type PTBTestWrapperRequest struct {
	Client      *iscmoveclient.Client
	Signer      cryptolib.Signer
	PackageID   iotago.PackageID
	GasPayments []*iotago.ObjectRef // optional
	GasPrice    uint64
	GasBudget   uint64
}

func PTBTestWrapper(
	req *PTBTestWrapperRequest,
	f func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	return req.Client.SignAndExecutePTB(
		context.TODO(),
		req.Signer,
		f(ptb).Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
	)
}
