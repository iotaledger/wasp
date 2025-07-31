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
	"io"
	"sync"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/util/pipe"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

// PeerMessageData is an envelope for all the messages exchanged via the peering module.
type PeerMessageData struct {
	PeeringID   PeeringID
	MsgReceiver byte
	MsgType     byte
	MsgData     []byte

	serializedData []byte
	serializedOnce sync.Once
}

func NewPeerMessageData(peeringID PeeringID, receiver byte, msgType byte, msgData []byte) *PeerMessageData {
	return &PeerMessageData{
		PeeringID:   peeringID,
		MsgReceiver: receiver,
		MsgType:     msgType,
		MsgData:     msgData,
	}
}

// peerMessageDataFromBytes creates a new PeerMessageData from bytes.
// The function takes ownership over "data" and the caller should not use "data" after this call.
func peerMessageDataFromBytes(data []byte) (ret *PeerMessageData, err error) {
	ret, err = rwutil.ReadFromBytes(data, new(PeerMessageData))
	if err != nil {
		return nil, err
	}
	ret.serializedOnce.Do(func() {
		ret.serializedData = data
	})
	return ret, nil
}

func (m *PeerMessageData) Bytes() []byte {
	m.serializedOnce.Do(func() {
		m.serializedData = rwutil.WriteToBytes(m)
	})
	return m.serializedData
}

func (m *PeerMessageData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	m.MsgReceiver = rr.ReadByte()
	m.MsgType = rr.ReadByte()
	rr.Read(&m.PeeringID)
	m.MsgData = rr.ReadBytes()
	return rr.Err
}

func (m *PeerMessageData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(m.MsgReceiver)
	ww.WriteByte(m.MsgType)
	ww.Write(&m.PeeringID)
	ww.WriteBytes(m.MsgData)
	return ww.Err
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
	peerMessageData, err := peerMessageDataFromBytes(data)
	if err != nil {
		return nil, err
	}
	return &PeerMessageNet{PeerMessageData: peerMessageData}, nil
}

func (m *PeerMessageNet) Bytes() []byte {
	return m.PeerMessageData.Bytes()
}

func (m *PeerMessageNet) GetHash() hashing.HashValue {
	m.hashOnce.Do(func() {
		m.hash = hashing.HashData(m.Bytes())
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
