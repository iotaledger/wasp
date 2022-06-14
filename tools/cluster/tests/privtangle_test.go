// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/nodeconn"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestHornetStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping privtangle test in short mode")
	}
	l1.StartPrivtangleIfNecessary(t.Logf)

	if l1.Privtangle == nil {
		t.Skip("tests running against live network, skipping pvt tangle tests")
	}
	// pvt tangle is already stated by the cluster l1_init
	ctx := context.Background()

	//
	// Try call the faucet.
	myKeyPair := cryptolib.NewKeyPair()
	myAddress := myKeyPair.GetPublicKey().AsEd25519Address()

	nc := nodeclient.New(l1.Config.APIAddress)
	nodeEvt, err := nc.EventAPI(ctx)
	require.NoError(t, err)
	require.NoError(t, nodeEvt.Connect(ctx))
	l1Info, err := nc.Info(ctx)
	require.NoError(t, err)

	myAddressOutputsCh, _ := nodeEvt.OutputsByUnlockConditionAndAddress(myAddress, l1Info.Protocol.Bech32HRP, nodeclient.UnlockConditionAny)

	log := testlogger.NewSilentLogger(t.Name(), true)
	client := nodeconn.NewL1Client(l1.Config, log)

	initialOutputCount := mustOutputCount(client, myAddress)
	//
	// Check if faucet requests are working.
	client.RequestFunds(myAddress)
	for i := 0; ; i++ {
		t.Logf("Waiting for a TX...")
		time.Sleep(100 * time.Millisecond)
		if initialOutputCount != mustOutputCount(client, myAddress) {
			break
		}
	}
	t.Logf("Waiting for output event...")
	outs := <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)

	//
	// Check if the TX post works.
	tx, err := nodeconn.MakeSimpleValueTX(client, l1.Config.FaucetKey, myAddress, 500_000)
	require.NoError(t, err)
	err = client.PostTx(tx)
	require.NoError(t, err)
	for i := 0; ; i++ {
		t.Logf("Waiting for a TX...")
		time.Sleep(100 * time.Millisecond)
		if initialOutputCount != mustOutputCount(client, myAddress) {
			break
		}
	}
	t.Logf("Waiting for output event...")
	outs = <-myAddressOutputsCh
	t.Logf("Waiting for output event, done: %+v", outs)
}

func mustOutputCount(client nodeconn.L1Client, myAddress *iotago.Ed25519Address) int {
	return len(mustOutputMap(client, myAddress))
}

func mustOutputMap(client nodeconn.L1Client, myAddress *iotago.Ed25519Address) map[iotago.OutputID]iotago.Output {
	outs, err := client.OutputMap(myAddress)
	if err != nil {
		panic(xerrors.Errorf("unable to get outputs as a map: %w", err))
	}
	return outs
}
