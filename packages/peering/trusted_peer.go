// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package peering provides an overlay network for communicating
// between nodes in a peer-to-peer style with low overhead
// encoding and persistent connections. The network provides only
// the asynchronous communication.
//
// It is intended to use for the committee consensus protocol.
//
package peering

import (
	"bytes"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/util"
)

// TrustedPeer carries a peer information we use to trust it.
type TrustedPeer struct {
	PubKey ed25519.PublicKey
	NetID  string
}

func TrustedPeerFromBytes(buf []byte) (*TrustedPeer, error) {
	var err error
	r := bytes.NewBuffer(buf)
	tp := TrustedPeer{}
	var keyBytes []byte
	if keyBytes, err = util.ReadBytes16(r); err != nil {
		return nil, err
	}
	tp.PubKey, _, err = ed25519.PublicKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	if tp.NetID, err = util.ReadString16(r); err != nil {
		return nil, err
	}
	return &tp, nil
}

func (tp *TrustedPeer) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := util.WriteBytes16(&buf, tp.PubKey.Bytes()); err != nil {
		return nil, err
	}
	if err := util.WriteString16(&buf, tp.NetID); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tp *TrustedPeer) PubKeyBytes() ([]byte, error) {
	return tp.PubKey.Bytes(), nil
}
