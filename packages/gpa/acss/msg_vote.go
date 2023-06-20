// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgVoteKind byte

const (
	msgVoteOK msgVoteKind = iota
	msgVoteREADY
)

// This message is used a vote for the "Bracha-style totality" agreement.
type msgVote struct {
	gpa.BasicMessage
	kind msgVoteKind
}

var _ gpa.Message = new(msgVote)

func (msg *msgVote) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeVote.ReadAndVerify(rr)
	msg.kind = msgVoteKind(rr.ReadByte())
	return rr.Err
}

func (msg *msgVote) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeVote.Write(ww)
	ww.WriteByte(byte(msg.kind))
	return ww.Err
}
