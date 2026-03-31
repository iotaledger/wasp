// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

// executed in cluster_test.go
func (e *ChainEnv) testEVMJsonRPCCluster(t *testing.T) {
	ctenv := newClusterTestEnv(t, e, 0)
	ctenv.TestRPCGetLogs()
	ctenv.TestRPCInvalidNonce()
	ctenv.TestRPCGasLimitTooLow()
	ctenv.TestRPCAccessHistoricalState()
	ctenv.TestGasPrice()
}

func TestEVMJsonRPCClusterAccessNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}
	clu := newCluster(t, waspClusterOpts{nNodes: 5})
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), []int{0, 1, 2, 3}, uint16(3))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)
	e := newClusterTestEnv(t, env, 4) // node #4 is an access node
	e.TestRPCGetLogs()
}

func TestEVMJsonRPCZeroGasFee(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cluster tests in short mode")
	}
	clu := newCluster(t, waspClusterOpts{nNodes: 5})
	chain, err := clu.DeployChainWithDKG(clu.Config.AllNodes(), []int{0, 1, 2, 3}, uint16(3))
	require.NoError(t, err)
	env := newChainEnv(t, clu, chain)
	e := newClusterTestEnv(t, env, 4) // node #4 is an access node

	fp1 := gas.DefaultFeePolicy()
	fp1.GasPerToken = util.Ratio32{
		A: 0,
		B: 0,
	}
	govClient := e.Chain.Client(e.Chain.OriginatorKeyPair)
	reqTx, err := govClient.PostRequest(context.Background(), governance.FuncSetFeePolicy.Message(fp1), chainclient.PostRequestParams{
		Transfer:  isc.NewAssets(iotaclient.DefaultGasBudget + 10),
		GasBudget: iotaclient.DefaultGasBudget,
	})
	require.NoError(t, err)
	_, err = e.Chain.CommitteeMultiClient().WaitUntilAllRequestsProcessedSuccessfully(context.Background(), e.Chain.ChainID, reqTx, false, 30*time.Second)
	require.NoError(t, err)

	d, err := govClient.CallView(context.Background(), governance.ViewGetFeePolicy.Message())
	require.NoError(t, err)
	fp2, err := governance.ViewGetFeePolicy.DecodeOutput(d)
	require.NoError(t, err)
	require.Equal(t, fp1, fp2)
	e.TestRPCGetLogs()
}
