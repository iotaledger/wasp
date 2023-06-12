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
	"sync"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// PeerMessage is an envelope for all the messages exchanged via the peering module.
type PeerMessageData struct {
	PeeringID   PeeringID
	MsgReceiver byte
	MsgType     byte
	MsgData     []byte

	serializedErr  error
	serializedData []byte
	serializedOnce sync.Once
}

// newPeerMessageDataFromBytes creates a new PeerMessageData from bytes.
// The function takes ownership over "data" and the caller should not use "data" after this call.
//

func newPeerMessageDataFromBytes(data []byte) (*PeerMessageData, error) {
	// create a copy of the slice for later usage of the raw data.
	cpy := lo.CopySlice(data)

	rr := rwutil.NewBytesReader(data)
	m := new(PeerMessageData)
	m.MsgReceiver = rr.ReadByte()
	m.MsgType = rr.ReadByte()
	rr.Read(&m.PeeringID)
	m.MsgData = rr.ReadBytes()
	if rr.Err != nil {
		return nil, rr.Err
	}

	m.serializedOnce.Do(func() {
		m.serializedErr = nil
		m.serializedData = cpy
	})

	return m, nil
}

func (m *PeerMessageData) Bytes() ([]byte, error) {
	m.serializedOnce.Do(func() {
		ww := rwutil.NewBytesWriter()
		ww.WriteByte(m.MsgReceiver)
		ww.WriteByte(m.MsgType)
		ww.Write(&m.PeeringID)
		ww.WriteBytes(m.MsgData)
		m.serializedData = ww.Bytes()
	})
	return m.serializedData, nil
}

type PeerMessageNet struct {
	*PeerMessageData

	hash     hashing.HashValue
	hashOnce sync.Once
}

var _ pipe.Hashable = &PeerMessageNet{}

// PeerMessageNetFromBytes creates a new PeerMessageNet from bytes.
// The function takes ownership over "data" and the caller should not use "data" after this call.
func PeerMessageNetFromBytes(data []byte) (*PeerMessageNet, error) {
	peerMessageData, err := newPeerMessageDataFromBytes(data)
	if err != nil {
		return nil, err
	}

	peerMessageNet := &PeerMessageNet{
		PeerMessageData: peerMessageData,
	}

	return peerMessageNet, nil
}

func (m *PeerMessageNet) Bytes() ([]byte, error) {
	return m.PeerMessageData.Bytes()
}

func (m *PeerMessageNet) GetHash() hashing.HashValue {
	m.hashOnce.Do(func() {
		bytes, err := m.Bytes()
		if err != nil {
			m.hash = hashing.HashValue{}
			return
		}

		m.hash = hashing.HashData(bytes)
	})

	return m.hash
}

type PeerMessageIn struct {
	*PeerMessageData
	SenderPubKey *cryptolib.PublicKey
}

type PeerMessageGroupIn struct {
	*PeerMessageIn
	SenderIndex uint16
}
