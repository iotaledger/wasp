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
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/testutil/privtangle"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

// TODO privtangle tests might conflict with cluster tests when running in parallel, check..

func TestHornetStartup(t *testing.T) {
	ctx := context.Background()
	tempDir := filepath.Join(os.TempDir(), "wasp-hornet-private_tangle")
	pt := privtangle.Start(ctx, tempDir, 16500, 3, t)

	//
	// Try call the faucet.
	myKeyPair := cryptolib.NewKeyPair()
	myAddress := myKeyPair.GetPublicKey().AsEd25519Address()

	nodeEvt := iotagox.NewNodeEventAPIClient(fmt.Sprintf("ws://localhost:%d/mqtt", pt.NodePortRestAPI(0)))
	require.NoError(t, nodeEvt.Connect(ctx))
	myAddressOutputsCh := nodeEvt.OutputsByUnlockConditionAndAddress(myAddress, iscp.NetworkPrefix, iotagox.UnlockConditionAny)

	initialOutputCount := mustOutputCount(pt, myAddress)

	client := cluster.NewL1Client(pt.L1Config())
	//
	// Check if faucet requests are working.
	client.(*cluster.L1Client).FaucetRequestHTTP(myAddress)
	for i := 0; ; i++ {
		t.Logf("Waiting for a TX...")
		time.Sleep(100 * time.Millisecond)
		if initialOutputCount != mustOutputCount(pt, myAddress) {
			break
		}
	}
	t.Logf("Waiting for output event...")
	outs := <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)

	//
	// Check if the TX post works.
	msg, err := client.(*cluster.L1Client).PostSimpleValueTX(pt.FaucetKeyPair, myAddress, 50000)
	require.NoError(t, err)
	t.Logf("Posted messageID=%v", msg.MustID())
	for i := 0; ; i++ {
		t.Logf("Waiting for a TX...")
		time.Sleep(100 * time.Millisecond)
		if initialOutputCount != mustOutputCount(pt, myAddress) {
			break
		}
	}
	t.Logf("Waiting for output event...")
	outs = <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)
}

func mustOutputCount(pt *privtangle.PrivTangle, myAddress *iotago.Ed25519Address) int {
	return len(mustOutputMap(pt, myAddress))
}

func mustOutputMap(pt *privtangle.PrivTangle, myAddress *iotago.Ed25519Address) map[iotago.OutputID]iotago.Output {
	client := cluster.NewL1Client(pt.L1Config())
	outs, err := client.OutputMap(myAddress)
	if err != nil {
		panic(xerrors.Errorf("unable to get outputs as a map: %w", err))
	}
	return outs
}
