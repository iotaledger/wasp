// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package semi implements a Common Coin (CC) that produces deterministic
// values only for some of the rounds. I.e. only every 5th round will give a
// real non-deterministic CC.
package semi

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type ccSemi struct {
	target gpa.GPA
	index  int
	output *bool
}

var _ gpa.GPA = &ccSemi{}

func New(index int, target gpa.GPA) gpa.GPA {
	return &ccSemi{index: index, target: target}
}

func (cc *ccSemi) Input(input gpa.Input) []*gpa.MessageOut {
	if input != nil {
		panic(errors.New("input must be nil"))
	}
	mod5 := cc.index % 5
	if mod5 < 2 {
		coin := true
		cc.output = &coin
		return nil
	}
	if mod5 < 4 {
		coin := false
		cc.output = &coin
		return nil
	}
	msgs := cc.target.Input(input)
	cc.checkOutput()
	return msgs
}

func (cc *ccSemi) Message(msg *gpa.MessageIn) []*gpa.MessageOut {
	if cc.output != nil {
		return nil
	}
	msgs := cc.target.Message(msg)
	cc.checkOutput()
	return msgs
}

func (cc *ccSemi) checkOutput() {
	if cc.output != nil {
		return
	}
	if out := cc.target.Output(); out != nil {
		cc.output = out.(*bool)
	}
}

func (cc *ccSemi) Output() gpa.Output {
	if cc.output == nil {
		return nil // Untyped nil.
	}
	return cc.output
}

func (cc *ccSemi) StatusString() string {
	if cc.output != nil {
		// Try produce compact output.
		return fmt.Sprintf("{CC:semi, index=%v, output=%v}", cc.index, *cc.output)
	}
	return fmt.Sprintf("{CC:semi, index=%v, output=%v, target=%v}", cc.index, cc.output, cc.target.StatusString())
}

func (cc *ccSemi) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return cc.target.UnmarshalPayload(data)
}
