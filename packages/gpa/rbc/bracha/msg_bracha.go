// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package bracha

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgBrachaType byte

const (
	msgBrachaTypePropose msgBrachaType = iota
	msgBrachaTypeEcho
	msgBrachaTypeReady
)

type msgBracha struct {
	gpa.BasicMessage
	brachaType msgBrachaType `bcs:""` // Type
	value      []byte        `bcs:""` // Value
}

var _ gpa.Message = new(msgBracha)
