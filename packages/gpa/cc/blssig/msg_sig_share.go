// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgSigShare struct {
	gpa.BasicMessage
	sigShare []byte `bcs:""`
}

var _ gpa.Message = new(msgSigShare)
