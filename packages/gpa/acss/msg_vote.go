// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgVoteKind byte

const (
	msgVoteOK msgVoteKind = iota
	msgVoteREADY
)

// This message is used a vote for the "Bracha-style totality" agreement.
type msgVote struct {
	gpa.BasicMessage
	kind msgVoteKind `bcs:""`
}

var _ gpa.Message = new(msgVote)
