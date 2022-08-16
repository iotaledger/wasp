// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

// import (
// 	"bytes"
// 	"io"

// 	iotago "github.com/iotaledger/iota.go/v3"
// 	"github.com/iotaledger/wasp/packages/hashing"
// 	"github.com/iotaledger/wasp/packages/tcrypto"
// 	"github.com/iotaledger/wasp/packages/util"
// 	"go.dedis.ch/kyber/v3/share"
// 	"go.dedis.ch/kyber/v3/sign/dss"
// )

// // Consensus -> Consensus
// type SignedResultMsg struct { // TODO: XXX: Remove it?
// 	ChainInputID *iotago.UTXOInput
// 	EssenceHash  hashing.HashValue
// 	SigShare     *dss.PartialSig
// }

// type SignedResultMsgIn struct { // TODO: XXX: Remove it?
// 	SignedResultMsg
// 	SenderIndex uint16
// }

// func NewSignedResultMsg(data []byte) (*SignedResultMsg, error) {
// 	msg := &SignedResultMsg{}
// 	r := bytes.NewReader(data)
// 	var err error
// 	if err := util.ReadHashValue(r, &msg.EssenceHash); err != nil {
// 		return nil, err
// 	}
// 	var uintTmp uint16
// 	//nolint:gocritic
// 	if err = util.ReadUint16(r, &uintTmp); err != nil {
// 		return nil, err
// 	}
// 	msg.SigShare = &dss.PartialSig{
// 		Partial: &share.PriShare{
// 			I: int(uintTmp),
// 			V: tcrypto.DefaultEd25519Suite().Scalar(),
// 		},
// 	}
// 	if _, err = msg.SigShare.Partial.V.UnmarshalFrom(r); err != nil {
// 		return nil, err
// 	}
// 	if msg.SigShare.SessionID, err = util.ReadBytes16(r); err != nil {
// 		return nil, err
// 	}
// 	if msg.SigShare.Signature, err = util.ReadBytes16(r); err != nil {
// 		return nil, err
// 	}
// 	if msg.ChainInputID, err = util.ReadOutputID(r); err != nil {
// 		return nil, err
// 	}
// 	return msg, nil
// }

// func (msg *SignedResultMsg) Write(w io.Writer) error {
// 	if _, err := w.Write(msg.EssenceHash[:]); err != nil {
// 		return err
// 	}
// 	if err := util.WriteUint16(w, uint16(msg.SigShare.Partial.I)); err != nil {
// 		return err
// 	}
// 	if _, err := msg.SigShare.Partial.V.MarshalTo(w); err != nil {
// 		return err
// 	}
// 	if err := util.WriteBytes16(w, msg.SigShare.SessionID); err != nil {
// 		return err
// 	}
// 	if err := util.WriteBytes16(w, msg.SigShare.Signature); err != nil {
// 		return err
// 	}
// 	if err := util.WriteOutputID(w, msg.ChainInputID); err != nil {
// 		return err
// 	}
// 	return nil
// }
