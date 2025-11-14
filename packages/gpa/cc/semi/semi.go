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
	"github.com/iotaledger/wasp/v2/packages/gpa/cc/blssig"
)

type CCSemi struct {
	target *blssig.CC
	index  int
	output *bool
}

func New(index int, target *blssig.CC) *CCSemi {
	return &CCSemi{index: index, target: target}
}

func (cc *CCSemi) Input(input gpa.Input) []gpa.PayloadOut {
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

func (cc *CCSemi) HandleMsgSigShare(msg gpa.PayloadIn[blssig.MsgSigShare]) []gpa.PayloadOut {
	if cc.output != nil {
		return nil
	}
	msgs := cc.target.HandleMsgSigShare(msg)
	cc.checkOutput()
	return msgs
}

func (cc *CCSemi) checkOutput() {
	if cc.output != nil {
		return
	}
	if out := cc.target.Output(); out != nil {
		cc.output = out.(*bool)
	}
}

func (cc *CCSemi) Output() gpa.Output {
	if cc.output == nil {
		return nil // Untyped nil.
	}
	return cc.output
}

func (cc *CCSemi) StatusString() string {
	if cc.output != nil {
		// Try produce compact output.
		return fmt.Sprintf("{CC:semi, index=%v, output=%v}", cc.index, *cc.output)
	}
	return fmt.Sprintf("{CC:semi, index=%v, output=%v, target=%v}", cc.index, cc.output, cc.target.StatusString())
}
