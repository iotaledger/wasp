// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"io"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type msgBrachaType byte

const (
	msgBrachaTypePropose msgBrachaType = iota
	msgBrachaTypeEcho
	msgBrachaTypeReady
)

type msgBracha struct {
	gpa.BasicMessage
	brachaType msgBrachaType // Type
	value      []byte        // Value
}

var _ gpa.Message = new(msgBracha)

func (msg *msgBracha) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msg.brachaType = msgBrachaType(rr.ReadByte())
	msg.value = rr.ReadBytes()
	return rr.Err
}

func (msg *msgBracha) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(byte(msg.brachaType))
	ww.WriteBytes(msg.value)
	return ww.Err
}
