// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"golang.org/x/xerrors"
)

type trustedNetworkManager struct {
	data map[ed25519.PublicKey]string
}

func NewTrustedNetworkManager() peering.TrustedNetworkManager {
	return &trustedNetworkManager{
		data: map[ed25519.PublicKey]string{},
	}
}

// IsTrustedPeer implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) IsTrustedPeer(pubKey ed25519.PublicKey) error {
	if _, ok := tnm.data[pubKey]; ok {
		return nil
	}
	return xerrors.New("not trusted")
}

// TrustPeer implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) TrustPeer(pubKey ed25519.PublicKey, netID string) (*peering.TrustedPeer, error) {
	tnm.data[pubKey] = netID
	return &peering.TrustedPeer{PubKey: pubKey, NetID: netID}, nil
}

// DistrustPeer implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) DistrustPeer(pubKey ed25519.PublicKey) (*peering.TrustedPeer, error) {
	delete(tnm.data, pubKey)
	return nil, nil
}

// TrustedPeers implements the peering.TrustedNetworkManager interface.
func (tnm *trustedNetworkManager) TrustedPeers() ([]*peering.TrustedPeer, error) {
	res := []*peering.TrustedPeer{}
	for k, v := range tnm.data {
		res = append(res, &peering.TrustedPeer{PubKey: k, NetID: v})
	}
	return res, nil
}
