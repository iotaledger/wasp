package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestMissingRequests(t *testing.T) {
	t.Skip("TODO: fix or remove test")

	clu := newCluster(t, waspClusterOpts{nNodes: 4})
	cmt := []int{0, 1, 2, 3}
	threshold := uint16(4)
	addr, err := clu.RunDKG(cmt, threshold)
	require.NoError(t, err)

	chain, err := clu.DeployChain(clu.Config.AllNodes(), cmt, threshold, addr)
	require.NoError(t, err)
	chainID := chain.ChainID

	chEnv := newChainEnv(t, clu, chain)

	userWallet, _, err := chEnv.Clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	// deposit funds before sending the off-ledger request
	chEnv.DepositFunds(iotaclient.DefaultGasBudget, userWallet)

	// TODO: Validate offleder logic
	// send off-ledger request to all nodes except #3
	req := isc.NewOffLedgerRequest(chainID, inccounter.FuncIncCounter.Message(nil), 0, gas.LimitsDefault.MaxGasPerRequest).Sign(userWallet)

	_, err = clu.WaspClient(0).RequestsAPI.OffLedger(context.Background()).OffLedgerRequest(apiclient.OffLedgerRequest{
		Request: cryptolib.EncodeHex(req.Bytes()),
	}).Execute()
	require.NoError(t, err)

	//------
	// send a dummy request to node #3, so that it proposes a batch and the consensus hang is broken
	req2 := isc.NewOffLedgerRequest(chainID, isc.NewMessageFromNames("foo", "bar"), 1, gas.LimitsDefault.MaxGasPerRequest).Sign(userWallet)

	_, err = clu.WaspClient(0).RequestsAPI.OffLedger(context.Background()).OffLedgerRequest(apiclient.OffLedgerRequest{
		Request: cryptolib.EncodeHex(req2.Bytes()),
	}).Execute()
	require.NoError(t, err)
	//-------

	// expect request to be successful, as node #3 must ask for the missing request from other nodes
	waitUntil(t, chEnv.counterEquals(43), clu.Config.AllNodes(), 30*time.Second)
}
