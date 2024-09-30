// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import "errors"

// silentNode can be used in the tests to model byzantine nodes, that
// are just consuming messages and not sending any messages at all.
type silentNode struct{}

var _ GPA = &silentNode{}

func MakeTestSilentNode() GPA {
	return &silentNode{}
}

func (s *silentNode) Input(input Input) OutMessages {
	return nil
}

func (s *silentNode) Message(msg Message) OutMessages {
	return nil
}

func (s *silentNode) Output() Output {
	return nil
}

func (s *silentNode) StatusString() string {
	return "{silentNode}"
}

func (s *silentNode) UnmarshalMessage(data []byte) (Message, error) {
	panic(errors.New("not implemented"))
}
