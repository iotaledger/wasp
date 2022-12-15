// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"fmt"
	"testing"

	"github.com/labstack/echo/v4"
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

// waspServicesMock is a mock implementation of the WaspServices interface
type waspServicesMock struct {
	solo   *solo.Solo
	chains map[isc.ChainID]*solo.Chain
}

var _ WaspServicesInterface = &waspServicesMock{}

func (w *waspServicesMock) ConfigDump() map[string]interface{} {
	return map[string]interface{}{
		"foo": "bar",
	}
}

func (*waspServicesMock) WebAPIPort() string {
	return "8080"
}

func (w *waspServicesMock) ExploreAddressBaseURL() string {
	return "http://127.0.0.1:8081/explorer/address"
}

func (w *waspServicesMock) PeeringStats() (*PeeringStats, error) {
	return &PeeringStats{
		Peers: []Peer{
			{
				NumUsers: 2,
				NetID:    "127.0.0.1:4001",
				IsAlive:  false,
			},
			{
				NumUsers: 2,
				NetID:    "127.0.0.1:4002",
				IsAlive:  false,
			},
			{
				NumUsers: 3,
				NetID:    "127.0.0.1:4002",
				IsAlive:  true,
			},
		},
		TrustedPeers: []TrustedPeer{
			{
				NetID:  "127.0.0.1:4000",
				PubKey: cryptolib.PublicKey{},
			},
			{
				NetID:  "127.0.0.1:4001",
				PubKey: cryptolib.PublicKey{},
			},
		},
	}, nil
}

func (w *waspServicesMock) MyNetworkID() string {
	return "127.0.0.1:4000"
}

func (w *waspServicesMock) ChainRecords() ([]*registry.ChainRecord, error) {
	var ret []*registry.ChainRecord
	for _, ch := range w.chains {
		chr, err := w.GetChainRecord(ch.ChainID)
		if err != nil {
			return nil, err
		}
		ret = append(ret, chr)
	}
	return ret, nil
}

func (w *waspServicesMock) GetChainRecord(chainID isc.ChainID) (*registry.ChainRecord, error) {
	return registry.NewChainRecord(chainID, true, []*cryptolib.PublicKey{}), nil
}

func (w *waspServicesMock) CallView(chainID isc.ChainID, scName, fname string, args dict.Dict) (dict.Dict, error) {
	ch, ok := w.chains[chainID]
	if !ok {
		return nil, xerrors.Errorf("chain not found")
	}
	return ch.CallView(scName, fname, args)
}

func (w *waspServicesMock) GetChainCommitteeInfo(chainID isc.ChainID) (*chain.CommitteeInfo, error) {
	_, ok := w.chains[chainID]
	if !ok {
		return nil, xerrors.Errorf("chain not found")
	}

	address := cryptolib.NewKeyPair().GetPublicKey().AsEd25519Address()

	return &chain.CommitteeInfo{
		Address:       address,
		Size:          2,
		Quorum:        1,
		QuorumIsAlive: true,
		PeerStatus: []*chain.PeerStatus{
			{
				Index:     0,
				NetID:     "localhost:2000",
				PubKey:    cryptolib.NewKeyPair().GetPublicKey(),
				Connected: true,
			},
			{
				Index:     1,
				NetID:     "localhost:2001",
				PubKey:    cryptolib.NewKeyPair().GetPublicKey(),
				Connected: true,
			},
		},
	}, nil
}

func (w *waspServicesMock) GetChainNodeConnectionMetrics(isc.ChainID) (nodeconnmetrics.NodeConnectionMessagesMetrics, error) {
	panic("Not implemented")
}

func (w *waspServicesMock) GetNodeConnectionMetrics() (nodeconnmetrics.NodeConnectionMetrics, error) {
	panic("Not implemented")
}

func (w *waspServicesMock) GetChainConsensusWorkflowStatus(chainID isc.ChainID) (chain.ConsensusWorkflowStatus, error) {
	panic("Not implemented")
}

func (w *waspServicesMock) GetChainConsensusPipeMetrics(chainID isc.ChainID) (chain.ConsensusPipeMetrics, error) {
	panic("Not implemented")
}

type dashboardTestEnv struct {
	wasp      *waspServicesMock
	echo      *echo.Echo
	dashboard *Dashboard
	solo      *solo.Solo
}

func (e *dashboardTestEnv) newChain() *solo.Chain {
	kp, _ := e.solo.NewKeyPairWithFunds()
	ch, _, _ := e.solo.NewChainExt(kp, 0, fmt.Sprintf("mock chain %d", len(e.wasp.chains)))
	e.wasp.chains[ch.ChainID] = ch
	return ch
}

func initDashboardTest(t *testing.T) *dashboardTestEnv {
	e := echo.New()
	s := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
	})
	w := &waspServicesMock{
		solo:   s,
		chains: make(map[isc.ChainID]*solo.Chain),
	}
	d := New(testlogger.NewLogger(t), e, w)
	return &dashboardTestEnv{
		wasp:      w,
		echo:      e,
		dashboard: d,
		solo:      s,
	}
}
