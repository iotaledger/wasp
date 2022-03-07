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
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
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

	initialOutputCount := mustOutputCount(ctx, pt, node0, myAddress)

	//
	// Check if faucet requests are working.
	pt.PostFaucetRequest(ctx, myAddress, iotago.PrefixTestnet)
	for i := 0; ; i++ {
		t.Logf("Waiting for a TX...")
		time.Sleep(100 * time.Millisecond)
		if initialOutputCount != mustOutputCount(ctx, pt, node0, myAddress) {
			break
		}
	}
	t.Logf("Waiting for output event...")
	outs := <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)

	//
	// Check if the TX post works.
	msg, err := pt.PostSimpleValueTX(ctx, node0, &pt.FaucetKeyPair, myAddress, 50000)
	require.NoError(t, err)
	t.Logf("Posted messageID=%v", msg.MustID())
	for i := 0; ; i++ {
		t.Logf("Waiting for a TX...")
		time.Sleep(100 * time.Millisecond)
		if initialOutputCount != mustOutputCount(ctx, pt, node0, myAddress) {
			break
		}
	}
	t.Logf("Waiting for output event...")
	outs = <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)

	//
	// Close.
	pt.Stop()
}

func mustOutputCount(ctx context.Context, pt *privtangle.PrivTangle, node0 *nodeclient.Client, myAddress *iotago.Ed25519Address) int {
	return len(mustOutputMap(ctx, pt, node0, myAddress))
}

func mustOutputMap(ctx context.Context, pt *privtangle.PrivTangle, node0 *nodeclient.Client, myAddress *iotago.Ed25519Address) map[iotago.OutputID]iotago.Output {
	outs, err := pt.OutputMap(ctx, node0, myAddress)
	if err != nil {
		panic(xerrors.Errorf("unable to get outputs as a map: %w", err))
	}
	return outs
}
