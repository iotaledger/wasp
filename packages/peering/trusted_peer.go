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
	Name       string
	id         *ComparablePubKey
	PeeringURL string
}

func NewTrustedPeer(name string, pubKey *cryptolib.PublicKey, peeringURL string) *TrustedPeer {
	return &TrustedPeer{
		Name:       name,
		id:         NewComparablePubKey(pubKey),
		PeeringURL: peeringURL,
	}
}

func (tp *TrustedPeer) ID() *ComparablePubKey {
	return tp.id
}

func (tp *TrustedPeer) Clone() onchangemap.Item[string, *ComparablePubKey] {
	return &TrustedPeer{
		Name:       tp.Name,
		id:         NewComparablePubKey(tp.PubKey().Clone()),
		PeeringURL: tp.PeeringURL,
	}
}

func (tp *TrustedPeer) PubKey() *cryptolib.PublicKey {
	return tp.ID().PubKey()
}

type jsonTrustedPeer struct {
	Name       string `json:"name"`
	PubKey     string `json:"publicKey"`
	PeeringURL string `json:"peeringURL"`
}

func (tp *TrustedPeer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonTrustedPeer{
		Name:       tp.Name,
		PubKey:     tp.PubKey().String(),
		PeeringURL: tp.PeeringURL,
	})
}

func (tp *TrustedPeer) UnmarshalJSON(bytes []byte) error {
	j := &jsonTrustedPeer{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}

	nodePubKey, err := cryptolib.NewPublicKeyFromString(j.PubKey)
	if err != nil {
		return err
	}

	*tp = *NewTrustedPeer(j.Name, nodePubKey, j.PeeringURL)

	return nil
}
