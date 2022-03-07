// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package privtangle_test

import (
	"context"
	"fmt"
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

	nodeEvt := iotagox.NewNodeEventAPIClient(fmt.Sprintf("ws://localhost:%d/mqtt", pt.NodePortRestAPI(0)))
	require.NoError(t, nodeEvt.Connect(ctx))
	myAddressOutputsCh := nodeEvt.OutputsByUnlockConditionAndAddress(myAddress, iotago.PrefixTestnet, iotagox.UnlockConditionAny)

	initialOutputCount := outputCount(ctx, t, node0, myAddress)

	if false {
		pt.PostFaucetRequest(ctx, myAddress, iotago.PrefixTestnet)

		for i := 0; ; i++ {
			time.Sleep(100 * time.Millisecond)
			if initialOutputCount != outputCount(ctx, t, node0, myAddress) {
				break
			}
		}
	} else {
		msg, err := pt.PostSimpleValueTX(ctx, node0, &pt.FaucetKeyPair, myAddress, 50000)
		require.NoError(t, err)
		t.Logf("Posted messageID=%v", msg.MustID())
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

	t.Logf("Waiting for output event...")
	outs := <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)

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
