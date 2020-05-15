package peering

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

// structure of the encoded PeerMessage:
// Timestamp   8 bytes
// MsgType type    1 byte
//  -- if MsgType == 0 (heartbeat) --> the end of message
//  -- if MsgType == 1 (handshake)
// MsgData (a string of peer network location) --> end of message
//  -- if MsgType >= FirstCommitteeMsgCode
// Address 32 bytes
// SenderIndex 2 bytes
// MsgData variable bytes to the end
//  -- otherwise panicL wrong MsgType

// always puts timestamp into first 8 bytes and 1 byte msg type
func encodeMessage(msg *PeerMessage) ([]byte, time.Time) {
	var buf bytes.Buffer
	// puts timestamp first
	ts := time.Now()
	_ = util.WriteUint64(&buf, uint64(ts.UnixNano()))
	switch {
	case msg == nil:
		buf.WriteByte(MsgTypeHeartbeat)

	case msg.MsgType == MsgTypeHeartbeat:
		buf.WriteByte(MsgTypeHeartbeat)

	case msg.MsgType == MsgTypeHandshake:
		buf.WriteByte(MsgTypeHandshake)
		buf.Write(msg.MsgData)

	case msg.MsgType >= FirstCommitteeMsgCode:
		buf.WriteByte(msg.MsgType)
		buf.Write(msg.Address.Bytes())
		_ = util.WriteUint16(&buf, msg.SenderIndex)
		_ = util.WriteBytes32(&buf, msg.MsgData)

	default:
		log.Panicf("wrong msg type %d", msg.MsgType)
	}
	return buf.Bytes(), ts
}

func decodeMessage(data []byte) (*PeerMessage, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("too short message")
	}
	rdr := bytes.NewBuffer(data)
	var uts uint64
	err := util.ReadUint64(rdr, &uts)
	if err != nil {
		return nil, err
	}
	ret := &PeerMessage{
		Timestamp: int64(uts),
	}
	ret.MsgType, err = util.ReadByte(rdr)
	if err != nil {
		return nil, err
	}
	switch {
	case ret.MsgType == MsgTypeHeartbeat:
		return ret, nil

	case ret.MsgType == MsgTypeHandshake:
		ret.MsgData = rdr.Bytes()
		return ret, nil

	case ret.MsgType >= FirstCommitteeMsgCode:
		// committee message
		if err = util.ReadAddress(rdr, &ret.Address); err != nil {
			return nil, err
		}
		if err = util.ReadUint16(rdr, &ret.SenderIndex); err != nil {
			return nil, err
		}
		if ret.MsgData, err = util.ReadBytes32(rdr); err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, fmt.Errorf("wrong message type %d", ret.MsgType)
}
