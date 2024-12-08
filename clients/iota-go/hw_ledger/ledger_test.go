package hw_ledger

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"

	ledger_go "github.com/iotaledger/wasp/clients/iota-go/hw_ledger/ledger-go"
)

func TestGetVersion(t *testing.T) {
	spec, err := ledger_go.NewSpeculosTransport(
		ledger_go.SpeculosTransportOpts{
			Host:     "localhost",
			ApduPort: 40000,
		},
	)
	require.NoError(t, err)

	dev := NewHWLedger(spec)
	require.NoError(t, err)

	version, err := dev.GetVersion()
	require.NoError(t, err)

	fmt.Println(version)

	require.Equal(t, version.Name, "iota")
}

func TestGetPublicKey(t *testing.T) {
	spec, err := ledger_go.NewSpeculosTransport(
		ledger_go.SpeculosTransportOpts{
			Host:     "localhost",
			ApduPort: 40000,
		},
	)
	require.NoError(t, err)

	dev := NewHWLedger(spec)
	require.NoError(t, err)

	version, err := dev.GetPublicKey("44'/4218'/0'/0'/0'", false) // true would require user interaction on the Ledger
	require.NoError(t, err)

	fmt.Println(version)
}

func TestSign(t *testing.T) {
	spec, err := ledger_go.NewSpeculosTransport(
		ledger_go.SpeculosTransportOpts{
			Host:     "localhost",
			ApduPort: 40000,
		},
	)
	require.NoError(t, err)

	dev := NewHWLedger(spec)
	require.NoError(t, err)

	address, err := dev.GetPublicKey("44'/4218'/0'/0'/0'", true)
	require.NoError(t, err)

	client := clients.NewL1Client(
		clients.L1Config{
			APIURL: iotaconn.AlphanetEndpointURL,
		},
	)

	fmt.Println(address)

	txnBytes, err := client.Publish(
		context.Background(),
		iotaclient.PublishRequest{
			Sender:          iotago.AddressFromArray([32]byte(address.Address)),
			CompiledModules: contracts.Testcoin().Modules,
			Dependencies:    contracts.Testcoin().Dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 5),
		},
	)
	require.NoError(t, err)

	version, err := dev.SignTransaction(
		"44'/4218'/0'/0'/0'",
		txnBytes.TxBytes,
	)
	require.NoError(t, err)

	fmt.Println(version)
}
