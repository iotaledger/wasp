// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/labstack/echo/v4"
	"golang.org/x/xerrors"
)

// waspServicesMock is a mock implementation of the WaspServices interface
type waspServicesMock struct {
	solo   *solo.Solo
	chains map[[ledgerstate.AddressLength]byte]*solo.Chain
}

func (w *waspServicesMock) ConfigDump() map[string]interface{} {
	return map[string]interface{}{
		"foo": "bar",
	}
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
				PubKey: [32]byte{},
			},
			{
				NetID:  "127.0.0.1:4001",
				PubKey: [32]byte{},
			},
		},
	}, nil
}

func (w *waspServicesMock) MyNetworkID() string {
	return "127.0.0.1:4000"
}

func (w *waspServicesMock) GetChainRecords() ([]*registry.ChainRecord, error) {
	var ret []*registry.ChainRecord
	for _, ch := range w.chains {
		chr, err := w.GetChainRecord(&ch.ChainID)
		if err != nil {
			return nil, err
		}
		ret = append(ret, chr)
	}
	return ret, nil
}

func (w *waspServicesMock) GetChainRecord(chainID *iscp.ChainID) (*registry.ChainRecord, error) {
	return &registry.ChainRecord{
		ChainID: chainID,
		Active:  true,
	}, nil
}

func (w *waspServicesMock) CallView(chainID *iscp.ChainID, scName, fname string, args dict.Dict) (dict.Dict, error) {
	ch, ok := w.chains[chainID.Array()]
	if !ok {
		return nil, xerrors.Errorf("chain not found")
	}
	return ch.CallView(scName, fname, args)
}

func (w *waspServicesMock) GetChainCommitteeInfo(chainID *iscp.ChainID) (*chain.CommitteeInfo, error) {
	_, ok := w.chains[chainID.Array()]
	if !ok {
		return nil, xerrors.Errorf("chain not found")
	}
	return &chain.CommitteeInfo{
		Address:       ledgerstate.NewED25519Address(ed25519.PublicKey{}),
		Size:          2,
		Quorum:        1,
		QuorumIsAlive: true,
		PeerStatus: []*chain.PeerStatus{
			{
				Index:     0,
				PeeringID: "0",
				IsSelf:    true,
				Connected: true,
			},
			{
				Index:     1,
				PeeringID: "1",
				IsSelf:    false,
				Connected: true,
			},
		},
	}, nil
}

type dashboardTestEnv struct {
	wasp      *waspServicesMock
	echo      *echo.Echo
	dashboard *Dashboard
	solo      *solo.Solo
}

func (e *dashboardTestEnv) newChain() *solo.Chain {
	ch := e.solo.NewChain(nil, fmt.Sprintf("mock chain %d", len(e.wasp.chains)))
	e.wasp.chains[ch.ChainID.Array()] = ch
	return ch
}

func initDashboardTest(t *testing.T) *dashboardTestEnv {
	e := echo.New()
	s := solo.New(t, false, true)
	w := &waspServicesMock{
		solo:   s,
		chains: make(map[[ledgerstate.AddressLength]byte]*solo.Chain),
	}
	d := Init(e, w, testlogger.NewLogger(t))
	return &dashboardTestEnv{
		wasp:      w,
		echo:      e,
		dashboard: d,
		solo:      s,
	}
}
