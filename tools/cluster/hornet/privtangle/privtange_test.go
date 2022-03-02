// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package privtangle_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	iotagob "github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/cluster/hornet/privtangle"
	"github.com/stretchr/testify/require"
)

// POST http://localhost:8091/api/plugins/faucet/v1/enqueue
// {"address":"atoi1qpszqzadsym6wpppd6z037dvlejmjuke7s24hm95s9fg9vpua7vluehe53e"}
func TestHornetStartup(t *testing.T) {
	ctx := context.Background()
	tempDir := filepath.Join(os.TempDir(), "wasp-hornet-private_tangle")
	pt := privtangle.Start(ctx, tempDir, 16500, 3, t)

	time.Sleep(3 * time.Second)

	//
	// Assert the node health.
	node0 := pt.NodeClient(0)
	health, err := node0.Health(ctx)
	require.NoError(t, err)
	require.True(t, health)

	//
	// Try call the faucet.
	myKeyPair := cryptolib.NewKeyPair()
	myAddress := cryptolib.Ed25519AddressFromPubKey(myKeyPair.PublicKey)
	faucetReq := fmt.Sprintf("{\"address\":%q}", myAddress.Bech32(iotago.PrefixTestnet))

	nodeEvt := iotagox.NewNodeEventAPIClient(fmt.Sprintf("ws://localhost:%d/mqtt", pt.NodePortRestAPI(0)))
	require.NoError(t, nodeEvt.Connect(ctx))
	// myAddressOutputsCh := nodeEvt.Ed25519AddressOutputs(myAddress) // TODO:
	// myAddressOutputsCh := nodeEvt.AddressOutputs(myAddress, iotago.PrefixTestnet)

	initialOutputCount := outputCount(ctx, t, node0, myAddress)

	if false {
		faucetURL := fmt.Sprintf("http://localhost:%d/api/plugins/faucet/v1/enqueue", pt.NodePortFaucet(0))
		httpReq, err := http.NewRequestWithContext(ctx, "POST", faucetURL, bytes.NewReader([]byte(faucetReq)))
		httpReq.Header.Set("Content-Type", "application/json")
		require.NoError(t, err)
		t.Logf("Calling faucet at %v", faucetURL)
		res, err := http.DefaultClient.Do(httpReq)
		require.NoError(t, err)
		resBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Logf("Response, status=%v, response=%s", res.Status, resBody)
		require.NoError(t, err)
		require.Equal(t, 202, res.StatusCode)

		for i := 0; ; i++ {
			time.Sleep(100 * time.Millisecond)
			if initialOutputCount != outputCount(ctx, t, node0, myAddress) {
				break
			}
		}
	} else {
		//
		// Build a TX.
		genesisAddr := cryptolib.Ed25519AddressFromPubKey(pt.FaucetKeyPair.PublicKey)
		genesisOuts := outputMap(ctx, t, node0, genesisAddr)
		var genesisOID iotago.OutputID
		var genesisOut iotago.Output
		for i, o := range genesisOuts {
			genesisOID = i
			genesisOut = o
			break
		}
		require.NotNil(t, genesisOID)
		require.NotNil(t, genesisOut)
		amount := uint64(50000)
		tx, err := iotagob.NewTransactionBuilder(
			iotago.NetworkIDFromString(pt.NetworkID),
		).AddInput(&iotagob.ToBeSignedUTXOInput{
			Address:  genesisAddr,
			OutputID: genesisOID,
			Output:   genesisOut,
		}).AddOutput(&iotago.BasicOutput{
			Amount:     amount,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: myAddress}},
		}).AddOutput(&iotago.BasicOutput{
			Amount:     genesisOut.Deposit() - amount,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: genesisAddr}},
		}).Build(
			iotago.ZeroRentParas,
			pt.FaucetKeyPair.AsAddressSigner(),
		)
		require.NoError(t, err)
		require.NotNil(t, tx)
		//
		// Build a message and post it.
		txMsg, err := iotagob.NewMessageBuilder().Payload(tx).Build()
		require.NoError(t, err)
		_, err = node0.SubmitMessage(ctx, txMsg)
		require.NoError(t, err)
		//
		// Wait for the TX to be approved.
		for i := 0; ; i++ {
			t.Logf("Waiting for a TX...")
			time.Sleep(100 * time.Millisecond)
			if initialOutputCount != outputCount(ctx, t, node0, myAddress) {
				break
			}
		}
	}

	// t.Logf("Waiting for output event...")
	// outs := <-myAddressOutputsCh
	// t.Logf("Waiting for output event, done: %+v", outs)

	//
	// Close.
	pt.Stop()
}

func outputCount(ctx context.Context, t *testing.T, node0 *nodeclient.Client, myAddress *iotago.Ed25519Address) int {
	return len(outputMap(ctx, t, node0, myAddress))
}
func outputMap(ctx context.Context, t *testing.T, node0 *nodeclient.Client, myAddress *iotago.Ed25519Address) map[iotago.OutputID]iotago.Output {
	res, err := node0.Indexer().Outputs(ctx, &nodeclient.OutputsQuery{
		AddressBech32: myAddress.Bech32(iotago.PrefixTestnet),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	result := make(map[iotago.OutputID]iotago.Output)
	for res.Next() {
		outs, err := res.Outputs()
		require.NoError(t, err)
		oids := res.Response.Items.MustOutputIDs()
		for i, o := range outs {
			result[oids[i]] = o
		}
	}
	return result
}
