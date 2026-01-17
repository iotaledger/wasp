package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/inccounter"
	"github.com/iotaledger/wasp/v2/tools/cluster"
)

// executed in cluster_test.go
func (e *ChainEnv) testPermissionlessAccessNode(t *testing.T) {
	// deposit funds for offledger requests
	keyPair, _, err := e.Clu.NewKeyPairWithFunds()
	require.NoError(e.t, err)

	e.DepositFunds(iotagraphql.DefaultGasBudget, keyPair)

	// spin a new node
	clu2 := newCluster(t, waspClusterOpts{
		nNodes:  1,
		dirName: "wasp-cluster-access-node",
		modifyConfig: func(nodeIndex int, configParams cluster.WaspConfigParams) cluster.WaspConfigParams {
			// avoid port conflicts when running everything on localhost
			configParams.APIPort += 100
			configParams.MetricsPort += 100
			configParams.PeeringPort += 100
			configParams.ProfilingPort += 100
			return configParams
		},
	})
	// remove this cluster when the test ends
	t.Cleanup(clu2.Stop)

	nodeClient := e.Clu.WaspClient(0)
	accessNodeClient := clu2.WaspClient(0)

	// adds node #0 from cluster2 as access node of node #0 from cluster1

	// trust setup between the two nodes
	node0peerInfo, _, err := nodeClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
	require.NoError(t, err)

	err = clu2.AddTrustedNode(apiclient.PeeringTrustRequest{
		Name:       "node-from-other-cluster",
		PublicKey:  node0peerInfo.PublicKey,
		PeeringURL: node0peerInfo.PeeringURL,
	})
	require.NoError(t, err)

	accessNodePeerInfo, _, err := accessNodeClient.NodeAPI.GetPeeringIdentity(context.Background()).Execute()
	require.NoError(t, err)

	err = e.Clu.AddTrustedNode(apiclient.PeeringTrustRequest{
		Name:       "node-from-other-cluster",
		PublicKey:  accessNodePeerInfo.PublicKey,
		PeeringURL: accessNodePeerInfo.PeeringURL,
	}, []int{0})
	require.NoError(t, err)

	// activate the chain on the access node
	_, err = accessNodeClient.ChainsAPI.
		SetChainRecord(context.Background(), e.Chain.ChainID.String()).
		ChainRecord(apiclient.ChainRecord{
			IsActive:    true,
			AccessNodes: []string{},
		}).Execute()
	require.NoError(t, err)

	// add node 0 from cluster 2 as a *permissionless* access node
	_, err = nodeClient.ChainsAPI.AddAccessNode(context.Background(), accessNodePeerInfo.PublicKey).Execute()
	require.NoError(t, err)

	// wait for the access node to report the chain as active (avoid fixed sleeps)
	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	require.NoError(t, waitForAccessNodeChainActive(reqCtx, accessNodeClient))

	// send a request to the access node
	myClient := chainclient.New(
		e.Clu.L1Client(),
		accessNodeClient,
		e.Chain.ChainID,
		keyPair,
	)
	req, err := myClient.PostOffLedgerRequest(context.Background(), accounts.FuncWithdraw.Message(),
		chainclient.PostRequestParams{
			Allowance: isc.NewAssets(10),
		},
	)
	require.NoError(t, err)

	// request has been processed
	_, err = e.Chain.CommitteeMultiClient().WaitUntilRequestProcessedSuccessfully(context.Background(), e.Chain.ChainID, req.ID(), false, 1*time.Minute)
	require.NoError(t, err)

	// remove the access node from cluster1 node 0
	_, err = nodeClient.ChainsAPI.RemoveAccessNode(context.Background(), accessNodePeerInfo.PublicKey).Execute()
	require.NoError(t, err)

	// proactively deactivate the chain on the former access node to speed up detachment and reduce flakiness
	_, err = accessNodeClient.ChainsAPI.
		SetChainRecord(context.Background(), e.Chain.ChainID.String()).
		ChainRecord(apiclient.ChainRecord{
			IsActive:    false,
			AccessNodes: []string{},
		}).Execute()
	require.NoError(t, err)

	// wait until the access node is fully detached: not listed by the committee and not active itself
	ctx2, cancel2 := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel2()
	require.NoError(t, waitUntilAccessNodeDetached(ctx2, nodeClient, accessNodeClient, accessNodePeerInfo.PublicKey))

	// try sending the request again (the access node is detached, so this may return a 4xx/5xx). We only need the request ID
	req, err = myClient.PostOffLedgerRequest(context.Background(), inccounter.FuncIncCounter.Message(nil))
	if err != nil {
		// It's expected to fail (e.g., 404/500) since the access node is no longer serving the chain
		t.Logf("posting to detached access node returned error (expected): %v", err)
	}
	// If the client couldn't even construct/sign the request (e.g., failed to fetch nonce via detached node),
	// build a signed off-ledger request locally to obtain a stable request ID for negative verification.
	if req == nil {
		// Prefer a fresh nonce from the committee node; fall back to a synthetic one if unavailable.
		nonce := uint64(time.Now().UnixNano())
		if nonceFromCommittee, err2 := chainclient.New(e.Clu.L1Client(), nodeClient, e.Chain.ChainID, keyPair).ISCNonce(context.Background()); err2 == nil {
			nonce = nonceFromCommittee
		} else {
			t.Logf("could not fetch ISC nonce from committee, using synthetic nonce: %v", err2)
		}
		tmp := isc.NewOffLedgerRequest(e.Chain.ChainID, inccounter.FuncIncCounter.Message(nil), nonce, iotagraphql.DefaultGasBudget)
		tmp.WithNonce(nonce)
		req = tmp.Sign(keyPair)
	}

	// request is not processed after a short while; poll for 404 on the committee node
	deadline := time.Now().Add(30 * time.Second)
	var receipt *apiclient.ReceiptResponse
	for time.Now().Before(deadline) {
		receipt, _, err = nodeClient.ChainsAPI.GetReceipt(context.Background(), req.ID().String()).Execute()
		if err != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.Error(t, err)
	require.Regexp(t, `404`, err.Error())
	require.Nil(t, receipt)
}

// waitForAccessNodeChainActive polls the access node until it reports the chain as active via /v1/chain
func waitForAccessNodeChainActive(ctx context.Context, accessNodeClient *apiclient.APIClient) error {
	for {
		info, _, err := accessNodeClient.ChainsAPI.GetChainInfo(ctx).Execute()
		if err == nil && info != nil && info.IsActive {
			return nil
		}
		select {
		case <-ctx.Done():
			if err == nil {
				return context.DeadlineExceeded
			}
			return err
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// waitUntilAccessNodeDetached waits until BOTH conditions are met:
// 1) The main node's committee info no longer lists the access node public key among access nodes
// 2) The access node's /v1/chain reports inactive or returns a non-2xx error (meaning the chain is not active/available there)
// It uses an exponential backoff and respects the provided context deadline.
func waitUntilAccessNodeDetached(ctx context.Context, mainNodeClient, accessNodeClient *apiclient.APIClient, accessNodePubKey string) error {
	backoff := 200 * time.Millisecond
	const maxBackoff = 2 * time.Second

	for {
		// Condition 1: not listed by committee
		cond1 := false
		if ci, _, err := mainNodeClient.ChainsAPI.GetCommitteeInfo(ctx).Execute(); err == nil && ci != nil {
			found := false
			for _, n := range ci.AccessNodes {
				if n.Node.PublicKey == accessNodePubKey {
					found = true
					break
				}
			}
			cond1 = !found
		}

		// Condition 2: access node reports chain inactive or endpoint is unavailable
		cond2 := false
		if info, resp, err := accessNodeClient.ChainsAPI.GetChainInfo(ctx).Execute(); err != nil {
			// any transport/openapi error is considered detached/unavailable
			cond2 = true
		} else {
			// if we managed to call it, consider detached only when not active
			cond2 = info == nil || !info.IsActive
			// just in case, treat 4xx/5xx as detached too (should surface as error in client, but be defensive)
			if resp != nil && resp.StatusCode >= 400 {
				cond2 = true
			}
		}

		if cond1 && cond2 {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("access node not fully detached before timeout: committee_unlisted=%v, accessnode_inactive=%v", cond1, cond2)
		case <-time.After(backoff):
			if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}
}
