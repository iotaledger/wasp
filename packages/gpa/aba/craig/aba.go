// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// TODO: That's Craig's "Good-Case-Coin-Free" ABA consensus.
package craig

import (
	"errors"

	"github.com/iotaledger/wasp/packages/gpa"
)

type abaImpl struct{}

var _ gpa.GPA = &abaImpl{}

func New() gpa.GPA {
	return nil
}

func (a *abaImpl) Input(input gpa.Input) gpa.OutMessages {
	return nil
}

func (a *abaImpl) Message(msg gpa.Message) gpa.OutMessages {
	return nil
}

func (a *abaImpl) Output() gpa.Output {
	return nil
}

func (a *abaImpl) StatusString() string {
	return "{ABA:Craig, TBD}"
}

func (a *abaImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return nil, errors.New("not implemented") // TODO: XXX: Impl.
}
