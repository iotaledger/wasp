package iscmoveclient_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
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
	client := iscmoveclienttest.NewHTTPClient()
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), iotaclient.PublishRequest{
		Sender:          cryptolibSigner.Address().AsIotaAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      cryptolib.SignerToIotaSigner(cryptolibSigner),
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	fmt.Println(txnResponse)
}

func TestBuildISCContract(t *testing.T) {
	fmt.Println("===========1")
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get the executable path: %v\n", err)
		return
	}
	fmt.Println("===========2")
	scriptDir := filepath.Dir(execPath)

	// Construct the relative path
	fmt.Println("===========3")
	relativePath := "clients/iota-go/contracts/isc/"
	fmt.Println("===========4")
	targetPath := filepath.Join(scriptDir, relativePath)

	// Change to the target directory
	fmt.Println("===========5")
	if err := os.Chdir(targetPath); err != nil {
		fmt.Printf("Failed to change directory to %s: %v\n", targetPath, err)
		return
	}

	// Define the command to run
	fmt.Println("===========6")
	cmd := exec.Command("sh", "-c", "iota move build --dump-bytecode-as-base64 > bytecode.json")

	// Run the command
	fmt.Println("===========7")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to execute command: %v\nOutput: %s\n", err, string(output))
	} else {
		fmt.Println("Command executed successfully. Output written to bytecode.json.")
	}
}
