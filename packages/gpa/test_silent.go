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

func (s *silentNode) Input(input Input) []MessageOut {
	return nil
}

func (s *silentNode) Message(msg MessageIn[any]) []MessageOut {
	return nil
}

func (s *silentNode) Output() Output {
	return nil
}

func (s *silentNode) StatusString() string {
	return "{silentNode}"
}

func (s *silentNode) MarshalPayload(payload any) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (s *silentNode) UnmarshalPayload(data []byte) (any, error) {
	return nil, errors.New("not implemented")
}
