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
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
)

// PeerMessage is an envelope for all the messages exchanged via
// the peering module.
type PeerMessageData struct {
	PeeringID   PeeringID
	MsgReceiver byte
	MsgType     byte
	MsgData     []byte
}

type PeerMessageNet struct {
	PeerMessageData
	serialized *[]byte
}

type PeerMessageIn struct {
	PeerMessageData
	SenderPubKey *ed25519.PublicKey
}

type PeerMessageGroupIn struct {
	PeerMessageIn
	SenderIndex uint16
}

var _ pipe.Hashable = &PeerMessageNet{}

//nolint:gocritic
func NewPeerMessageDataFromBytes(buf []byte) (*PeerMessageData, error) {
	var err error
	r := bytes.NewBuffer(buf)
	m := PeerMessageData{}
	if m.MsgReceiver, err = util.ReadByte(r); err != nil {
		return nil, err
	}
	if m.MsgType, err = util.ReadByte(r); err != nil {
		return nil, err
	}
	if err = m.PeeringID.Read(r); err != nil {
		return nil, err
	}
	if m.MsgData, err = util.ReadBytes32(r); err != nil {
		return nil, err
	}
	return &m, nil
}

func NewPeerMessageNetFromBytes(buf []byte) (*PeerMessageNet, error) {
	data, err := NewPeerMessageDataFromBytes(buf)
	if err != nil {
		return nil, err
	}
	return &PeerMessageNet{
		PeerMessageData: *data,
		serialized:      &buf,
	}, nil
}

func (m *PeerMessageNet) Bytes() ([]byte, error) {
	if m.serialized == nil {
		serialized, err := m.PeerMessageData.bytes()
		if err != nil {
			return nil, err
		}
		m.serialized = &serialized
	}
	return *(m.serialized), nil
}

func (m *PeerMessageData) bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := util.WriteByte(&buf, m.MsgReceiver); err != nil {
		return nil, err
	}
	if err := util.WriteByte(&buf, m.MsgType); err != nil {
		return nil, err
	}
	if err := m.PeeringID.Write(&buf); err != nil {
		return nil, err
	}
	if err := util.WriteBytes32(&buf, m.MsgData); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *PeerMessageNet) GetHash() hashing.HashValue {
	mBytes, err := m.Bytes()
	if err != nil {
		return hashing.HashValue{}
	}
	return hashing.HashData(mBytes)
}
