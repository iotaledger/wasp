// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"golang.org/x/xerrors"
)

// DomainWithFallback acts as a peering domain, but maintains 2 sets of peers,
// one for the normal operation (matches with the chain peers) and other for
// the fallback mode (include all the trusted peers).
type DomainWithFallback struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	net       peering.NetworkProvider
	dom       peering.PeerDomainProvider
	fallback  bool
	mainPeers []*cryptolib.PublicKey
	log       *logger.Logger
}

func NewDomainWithFallback(peeringID peering.PeeringID, net peering.NetworkProvider, log *logger.Logger) (*DomainWithFallback, error) {
	dom, err := net.PeerDomain(peeringID, make([]*cryptolib.PublicKey, 0))
	if err != nil {
		return nil, xerrors.Errorf("unable to allocate peer domain: %w", err)
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	df := DomainWithFallback{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		net:       net,
		dom:       dom,
		fallback:  false,
		mainPeers: make([]*cryptolib.PublicKey, 0),
		log:       log,
	}
	go df.run()
	return &df, nil
}

func (df *DomainWithFallback) run() {
	for {
		select {
		case <-df.ctx.Done():
			return
		case <-time.After(1 * time.Second):
			if df.fallback {
				// Some peers could be made trusted/untrusted. Some can go online/offline.
				// Thus we need to update the list of peers periodically.
				// That is only needed of the fallback list is in use.
				df.updateDomainPeers()
			}
		}
	}
}

// SetMainPeers updates the peer list as it is reported by the chain.
// We exclude the self peer here, because we use this to send messages to other nodes.
func (df *DomainWithFallback) SetMainPeers(peers []*cryptolib.PublicKey) {
	selfPubKey := df.net.Self().PubKey()
	otherPeers := make([]*cryptolib.PublicKey, 0)
	for i := range peers {
		if *peers[i] != *selfPubKey {
			otherPeers = append(otherPeers, peers[i])
		}
	}
	df.log.Debugf("SetMainPeers: number of other mainPeers: %v, self dropped=%v", len(otherPeers), len(peers) != len(otherPeers))
	df.mainPeers = otherPeers
	if !df.fallback {
		df.updateDomainPeers()
	}
}

func (df *DomainWithFallback) HaveMainPeers() bool {
	return len(df.mainPeers) > 0
}

func (df *DomainWithFallback) SetFallbackMode(fallback bool) {
	if df.fallback == fallback {
		return
	}
	df.log.Debugf("SetFallbackMode: fallback=%v", fallback)
	df.fallback = fallback
	df.updateDomainPeers()
}

func (df *DomainWithFallback) GetFallbackMode() bool {
	return df.fallback
}

func (df *DomainWithFallback) updateDomainPeers() {
	var peers []*cryptolib.PublicKey
	if df.fallback {
		selfPubKey := df.net.Self().PubKey()
		allTrusted := make([]*cryptolib.PublicKey, 0)
		for _, n := range df.net.PeerStatus() {
			if n.IsAlive() && *n.PubKey() != *selfPubKey {
				allTrusted = append(allTrusted, n.PubKey())
			}
		}
		peers = allTrusted
	} else {
		peers = df.mainPeers
	}
	df.log.Debugf("updateDomainPeers: in fallback=%v mode will use %v nodes.", df.fallback, len(peers))
	df.dom.UpdatePeers(peers)
}

func (df *DomainWithFallback) Attach(receiver byte, callback func(recv *peering.PeerMessageIn)) interface{} {
	return df.dom.Attach(receiver, callback)
}

func (df *DomainWithFallback) Detach(attachID interface{}) {
	df.dom.Detach(attachID)
}

func (df *DomainWithFallback) Close() {
	df.dom.Close()
	df.ctxCancel()
}

func (df *DomainWithFallback) GetRandomOtherPeers(upToNumPeers int) []*cryptolib.PublicKey {
	return df.dom.GetRandomOtherPeers(upToNumPeers)
}

func (df *DomainWithFallback) SendMsgByPubKey(pubKey *cryptolib.PublicKey, msgReceiver, msgType byte, msgData []byte) {
	df.dom.SendMsgByPubKey(pubKey, msgReceiver, msgType, msgData)
}
