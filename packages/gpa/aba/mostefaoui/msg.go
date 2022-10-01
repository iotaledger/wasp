// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeVote byte = iota
	msgTypeDone
	msgTypeWrapped
)

// Implements the gpa.GPA interface.
func (a *abaImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return nil, xerrors.Errorf("not implemented") // TODO: Impl. UnmarshalMessage
}
