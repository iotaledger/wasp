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

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		res, err := node0.Indexer().Outputs(ctx, &nodeclient.OutputsQuery{
			AddressBech32: myAddress.Bech32(iotago.PrefixTestnet),
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		t.Logf("Outputs...")
		for res.Next() {
			outs, err := res.Outputs()
			require.NoError(t, err)
			for i, o := range outs {
				t.Logf("Outputs: %v=%#v", res.Response.Items[i], o)
			}
		}
		t.Logf("Outputs... Done")
	}

	// t.Logf("Waiting for output event...")
	// outs := <-myAddressOutputsCh
	// t.Logf("Waiting for output event, done: %+v", outs)

	//
	// Close.
	time.Sleep(3 * time.Second)
	pt.Stop()
}
