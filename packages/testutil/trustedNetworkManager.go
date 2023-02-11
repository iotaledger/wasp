// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"errors"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
)

type trustedNetworkManager struct {
	data         map[cryptolib.PublicKeyKey]trustedNetworkDataEntry
	changeEvents *event.Event[[]*peering.TrustedPeer]
}

type trustedNetworkDataEntry struct {
	name       string
	key        *cryptolib.PublicKey
	peeringURL string
}

var _ peering.TrustedNetworkManager = &trustedNetworkManager{}

func NewTrustedNetworkManager() peering.TrustedNetworkManager {
	return &trustedNetworkManager{
		data:         map[cryptolib.PublicKeyKey]trustedNetworkDataEntry{},
		changeEvents: event.New[[]*peering.TrustedPeer](),
	}
}

// IsTrustedPeer implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) IsTrustedPeer(pubKey *cryptolib.PublicKey) error {
	if _, ok := tnm.data[pubKey.AsKey()]; ok {
		return nil
	}
	return errors.New("not trusted")
}

// TrustPeer implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) TrustPeer(name string, pubKey *cryptolib.PublicKey, peeringURL string) (*peering.TrustedPeer, error) {
	tnm.data[pubKey.AsKey()] = trustedNetworkDataEntry{name, pubKey, peeringURL}
	tnm.changeEvents.Trigger(tnm.mustTrustedPeers())
	return peering.NewTrustedPeer(name, pubKey, peeringURL), nil
}

// DistrustPeer implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error) {
	delete(tnm.data, pubKey.AsKey())
	tnm.changeEvents.Trigger(tnm.mustTrustedPeers())
	return nil, nil
}

// TrustedPeers implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) TrustedPeers() ([]*peering.TrustedPeer, error) {
	return tnm.mustTrustedPeers(), nil
}

func (tnm *trustedNetworkManager) mustTrustedPeers() []*peering.TrustedPeer {
	res := []*peering.TrustedPeer{}
	for _, v := range tnm.data {
		res = append(res, peering.NewTrustedPeer(v.name, v.key, v.peeringURL))
	}
	return res
}

func (tnm *trustedNetworkManager) TrustedPeersListener(callback func([]*peering.TrustedPeer)) context.CancelFunc {
	callback(tnm.mustTrustedPeers())
	closure := event.NewClosure(callback)
	tnm.changeEvents.Attach(closure)
	return func() { tnm.changeEvents.Detach(closure) }
}
