// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package peering provides an overlay network for communicating
// between nodes in a peer-to-peer style with low overhead
// encoding and persistent connections. The network provides only
// the asynchronous communication.
//
// It is intended to use for the committee consensus protocol.
package peering

import (
	"encoding/json"

	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type ComparablePubKey struct {
	pubKey *cryptolib.PublicKey
}

func NewComparablePubKey(pubKey *cryptolib.PublicKey) *ComparablePubKey {
	return &ComparablePubKey{
		pubKey: pubKey,
	}
}

func (c *ComparablePubKey) PubKey() *cryptolib.PublicKey {
	return c.pubKey
}

func (c *ComparablePubKey) Key() string {
	return iotago.EncodeHex(c.pubKey.AsBytes())
}

func (c *ComparablePubKey) String() string {
	return iotago.EncodeHex(c.pubKey.AsBytes())
}

// TrustedPeer carries a peer information we use to trust it.
type TrustedPeer struct {
	id    *ComparablePubKey
	NetID string
}

func NewTrustedPeer(pubKey *cryptolib.PublicKey, netID string) *TrustedPeer {
	return &TrustedPeer{
		id:    NewComparablePubKey(pubKey),
		NetID: netID,
	}
}

func (tp *TrustedPeer) ID() *ComparablePubKey {
	return tp.id
}

func (tp *TrustedPeer) Clone() onchangemap.Item[string, *ComparablePubKey] {
	return &TrustedPeer{
		id:    NewComparablePubKey(tp.PubKey().Clone()),
		NetID: tp.NetID,
	}
}

func (tp *TrustedPeer) PubKey() *cryptolib.PublicKey {
	return tp.ID().PubKey()
}

type jsonTrustedPeer struct {
	PubKey string `json:"publicKey"`
	NetID  string `json:"netID"`
}

func (tp *TrustedPeer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonTrustedPeer{
		PubKey: cryptolib.PublicKeyToHex(tp.PubKey()),
		NetID:  tp.NetID,
	})
}

func (tp *TrustedPeer) UnmarshalJSON(bytes []byte) error {
	j := &jsonTrustedPeer{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}

	nodePubKey, err := cryptolib.NewPublicKeyFromHex(j.PubKey)
	if err != nil {
		return err
	}

	*tp = *NewTrustedPeer(nodePubKey, j.NetID)

	return nil
}
