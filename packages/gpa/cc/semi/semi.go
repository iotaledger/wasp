// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// semi package implements a Common Coin (CC) that produces deterministic
// values only for some of the rounds. I.e. only every 5th round will give a
// real non-deterministic CC.
package semi

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
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

func (cc *ccSemi) Input(input gpa.Input) gpa.OutMessages {
	if input != nil {
		panic(fmt.Errorf("input must be nil"))
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
	return cc.checkOutput(cc.target.Input(input))
}

func (cc *ccSemi) Message(msg gpa.Message) gpa.OutMessages {
	if cc.output != nil {
		return nil
	}
	return cc.checkOutput(cc.target.Message(msg))
}

func (cc *ccSemi) checkOutput(msgs gpa.OutMessages) gpa.OutMessages {
	if cc.output != nil {
		return msgs
	}
	out := cc.target.Output()
	if out != nil {
		cc.output = out.(*bool)
	}
	return msgs
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

func (cc *ccSemi) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return cc.target.UnmarshalMessage(data)
}
