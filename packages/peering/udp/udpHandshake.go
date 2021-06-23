// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package udp

import (
	"bytes"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type handshakeMsg struct {
	netID   string            // Their NetID
	pubKey  ed25519.PublicKey // Our PubKey.
	respond bool              // Do the message asks for a response?
}

func (m *handshakeMsg) bytes(secKey ed25519.PrivateKey) ([]byte, error) {
	var err error
	//
	// Payload.
	var payloadBuf bytes.Buffer
	if err = util.WriteString16(&payloadBuf, m.netID); err != nil {
		return nil, err
	}
	if err = util.WriteBytes16(&payloadBuf, m.pubKey.Bytes()); err != nil {
		return nil, err
	}
	if err = util.WriteBoolByte(&payloadBuf, m.respond); err != nil {
		return nil, err
	}
	//
	// Signed frame.
	payload := payloadBuf.Bytes()
	signature := secKey.Sign(payload)
	signedBuf := bytes.Buffer{}
	if err = util.WriteBytes16(&signedBuf, signature.Bytes()); err != nil {
		return nil, err
	}
	if err = util.WriteBytes16(&signedBuf, payload); err != nil {
		return nil, err
	}
	return signedBuf.Bytes(), nil
}

func handshakeMsgFromBytes(buf []byte) (*handshakeMsg, error) {
	var err error
	//
	// Signed frame.
	rSigned := bytes.NewReader(buf)
	var payload []byte
	var signatureBytes []byte
	if signatureBytes, err = util.ReadBytes16(rSigned); err != nil {
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
	var pubKeyBytes []byte
	if pubKeyBytes, err = util.ReadBytes16(rPayload); err != nil {
		return nil, err
	}
	if m.pubKey, _, err = ed25519.PublicKeyFromBytes(pubKeyBytes); err != nil {
		return nil, err
	}
	if err = util.ReadBoolByte(rPayload, &m.respond); err != nil {
		return nil, err
	}
	//
	// Verify the signature.
	var signature ed25519.Signature
	if signature, _, err = ed25519.SignatureFromBytes(signatureBytes); err != nil {
		return nil, err
	}
	if !m.pubKey.VerifySignature(payload, signature) {
		return nil, xerrors.New("invalid message signature")
	}
	return &m, nil
}
