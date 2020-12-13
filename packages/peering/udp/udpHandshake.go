package udp

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/bls"
)

type handshakeMsg struct {
	netID   string      // Their NetID
	pubKey  kyber.Point // Our PubKey.
	respond bool        // Do the message asks for a response?
}

func (m *handshakeMsg) bytes(secKey kyber.Scalar, suite Suite) ([]byte, error) {
	var err error
	//
	// Payload.
	var payloadBuf bytes.Buffer
	if err = util.WriteString16(&payloadBuf, m.netID); err != nil {
		return nil, err
	}
	if err = util.WriteMarshaled(&payloadBuf, m.pubKey); err != nil {
		return nil, err
	}
	if err = util.WriteBoolByte(&payloadBuf, m.respond); err != nil {
		return nil, err
	}
	var payload = payloadBuf.Bytes()
	var signature []byte
	if signature, err = bls.Sign(suite, secKey, payload); err != nil {
		return nil, err
	}
	//
	// Signed frame.
	var signedBuf bytes.Buffer
	if err = util.WriteBytes16(&signedBuf, signature); err != nil {
		return nil, err
	}
	if err = util.WriteBytes16(&signedBuf, payload); err != nil {
		return nil, err
	}
	return signedBuf.Bytes(), nil
}

func handshakeMsgFromBytes(buf []byte, suite Suite) (*handshakeMsg, error) {
	var err error
	//
	// Signed frame.
	rSigned := bytes.NewReader(buf)
	var payload []byte
	var signature []byte
	if signature, err = util.ReadBytes16(rSigned); err != nil {
		return nil, err
	}
	if payload, err = util.ReadBytes16(rSigned); err != nil {
		return nil, err
	}
	//
	// Payload.
	rPayload := bytes.NewReader(payload)
	m := handshakeMsg{}
	if m.netID, err = util.ReadString16(rPayload); err != nil {
		return nil, err
	}
	m.pubKey = suite.Point()
	if err = util.ReadMarshaled(rPayload, m.pubKey); err != nil {
		return nil, err
	}
	if err = util.ReadBoolByte(rPayload, &m.respond); err != nil {
		return nil, err
	}
	//
	// Verify the signature.
	if err = bls.Verify(suite, m.pubKey, payload, signature); err != nil {
		return nil, err
	}
	return &m, nil
}
