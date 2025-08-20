//go:build ledger_speculos
// +build ledger_speculos

package hw_ledger

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	ledger_go "github.com/iotaledger/wasp/v2/clients/iota-go/hw_ledger/ledger-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

func initializeLedger(t *testing.T) *HWLedger {
	debug := false

	var ledgerDevice *HWLedger
	var err error

	if debug {
		dev, err := ledger_go.NewSpeculosTransport(ledger_go.DefaultSpeculosTransportOpts())

		require.NoError(t, err)
		ledgerDevice = NewHWLedger(dev)

	} else {
		ledgerDevice, err = TryAndConnect()
		require.NoError(t, err)
	}

	return ledgerDevice
}

func TestGetVersion(t *testing.T) {
	dev := initializeLedger(t)

	version, err := dev.GetVersion()
	require.NoError(t, err)

	fmt.Println(version)

	require.Equal(t, version.Name, "iota")
}

func TestGetPublicKey(t *testing.T) {
	dev := initializeLedger(t)

	version, err := dev.GetPublicKey("44'/4218'/0'/0'/0'", false) // true would require user interaction on the Ledger
	require.NoError(t, err)

	fmt.Println(version)
}

func TestDeployChain(t *testing.T) {
	dev := initializeLedger(t)

	l1 := clients.NewL1Client(
		clients.L1Config{
			APIURL:    iotaconn.AlphanetEndpointURL,
			FaucetURL: iotaconn.AlphanetFaucetURL,
		},
	)

	pubKey, err := dev.GetPublicKey("44'/4218'/123'/0'/0'", false)
	require.NoError(t, err)

	err = l1.RequestFunds(context.Background(), pubKey.Address)
	require.NoError(t, err)

	signer := NewLedgerSigner(dev, "44'/4218'/123'/0'/0'", false)
	result, err := l1.DeployISCContracts(context.Background(), signer)
	require.NoError(t, err)

	fmt.Println(result)

}

func TestSign(t *testing.T) {
	dev := initializeLedger(t)

	someTX, _ := hexutil.Decode(
		"0x000000000002000840420f000000000000204f2370b2a4810ad6c8e1cfd92cc8c8818fef8f59e3a80cea17871f78d850ba4b0202000101000001010200000101006fb21feead027da4873295affd6c4f3618fe176fa2fbf3e7b5ef1d9463b31e210112a6d0c44edc630d2724b1f57fea4f93308b1d22164402c65778bd99379c4733070000000000000020f2fd3c87b227f1015182fe4348ed680d7ed32bcd3269704252c03e1d0b13d30d6fb21feead027da4873295affd6c4f3618fe176fa2fbf3e7b5ef1d9463b31e2101000000000000000c0400000000000000",
	)

	version, err := dev.SignTransaction(
		"44'/4218'/123'/0'/0'",
		iotasigner.MessageWithIntent(iotasigner.DefaultIntent(), someTX),
	)
	require.NoError(t, err)

	fmt.Println(version)
}
