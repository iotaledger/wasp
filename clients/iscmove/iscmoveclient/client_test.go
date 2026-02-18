package iscmoveclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
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
) (*iotagraphql.ExecuteTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	return req.Client.SignAndExecutePTB(
		context.Background(),
		req.Signer,
		f(ptb).Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
	)
}

func TestKeys(t *testing.T) {
	cryptolibSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	client := iscmoveclienttest.NewClient()
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), iotagraphql.PublishRequest{
		Sender:          *cryptolibSigner.Address().AsIotaAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget * 10),
	})
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes.TxBytes,
		cryptolib.SignerToIotaSigner(cryptolibSigner),
	)
	require.NoError(t, err)
	fmt.Println(txnResponse)
}
