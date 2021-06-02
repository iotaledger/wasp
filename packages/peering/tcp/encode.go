// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcp

import (
	"bytes"
	"fmt"
	"log"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
)

// structure of the encoded PeerMessage:
// OutputTimestamp   8 bytes
// MsgType type    1 byte
//  -- if MsgType == 0 (heartbeat) --> the end of message
//  -- if MsgType == 1 (handshake)
// MsgData (handshakeMsg) --> end of message
//  -- if MsgType >= FirstUserMsgCode
// PeeringID 32 bytes
// SenderIndex 2 bytes
// MsgData variable bytes to the end
//  -- otherwise panic wrong MsgType

const chunkMessageOverhead = 8 + 1

// always puts timestamp into first 8 bytes and 1 byte msg type
func encodeMessage(msg *peering.PeerMessage, ts int64) []byte {
	var buf bytes.Buffer
	// puts timestamp first
	_ = util.WriteUint64(&buf, uint64(ts))
	switch {
	case msg == nil:
		panic("msgTypeReserved")

	case msg.MsgType == msgTypeReserved:
		panic("msgTypeReserved")

	case msg.MsgType == msgTypeHandshake:
		buf.WriteByte(msgTypeHandshake)
		buf.Write(msg.MsgData)

	case msg.MsgType == msgTypeMsgChunk:
		buf.WriteByte(msgTypeMsgChunk)
		buf.Write(msg.MsgData)

	case msg.MsgType >= peering.FirstUserMsgCode:
		buf.WriteByte(msg.MsgType)
		msg.PeeringID.Write(&buf)
		util.WriteUint16(&buf, msg.SenderIndex)
		util.WriteBytes32(&buf, msg.MsgData)

	default:
		log.Panicf("wrong msg type %d", msg.MsgType)
	}
	return buf.Bytes()
}

func decodeMessage(data []byte) (*peering.PeerMessage, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("too short message")
	}
	rdr := bytes.NewBuffer(data)
	var uts uint64
	err := util.ReadUint64(rdr, &uts)
	if err != nil {
		return nil, err
	}
	ret := &peering.PeerMessage{
		Timestamp: int64(uts),
	}
	ret.MsgType, err = util.ReadByte(rdr)
	if err != nil {
		return nil, err
	}
	switch {
	case ret.MsgType == msgTypeHandshake:
		ret.MsgData = rdr.Bytes()
		return ret, nil

	case ret.MsgType == msgTypeMsgChunk:
		ret.MsgData = rdr.Bytes()
		return ret, nil

	case ret.MsgType >= peering.FirstUserMsgCode:
		// committee message
		if err = ret.PeeringID.Read(rdr); err != nil {
			return nil, err
		}
		if err = util.ReadUint16(rdr, &ret.SenderIndex); err != nil {
			return nil, err
		}
		if ret.MsgData, err = util.ReadBytes32(rdr); err != nil {
			return nil, err
		}
		return ret, nil

	default:
		return nil, fmt.Errorf("peering.decodeMessage.wrong message type: %d", ret.MsgType)
	}
}

type handshakeMsg struct {
	peeringID string      // Pair of peer NetIDs
	srcNetID  string      // Their NetID
	pubKey    kyber.Point // Our PubKey.
}

func (m *handshakeMsg) bytes() ([]byte, error) {
	var err error
	var buf bytes.Buffer
	if err = util.WriteString16(&buf, m.peeringID); err != nil {
		return nil, err
	}
	if err = util.WriteString16(&buf, m.srcNetID); err != nil {
		return nil, err
	}
	if err = util.WriteMarshaled(&buf, m.pubKey); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func handshakeMsgFromBytes(buf []byte, suite kyber.Group) (*handshakeMsg, error) {
	var err error
	r := bytes.NewReader(buf)
	m := handshakeMsg{}
	if m.peeringID, err = util.ReadString16(r); err != nil {
		return nil, err
	}
	if m.srcNetID, err = util.ReadString16(r); err != nil {
		return nil, err
	}
	m.pubKey = suite.Point()
	if err = util.ReadMarshaled(r, m.pubKey); err != nil {
		return nil, err
	}
	return &m, nil
}
