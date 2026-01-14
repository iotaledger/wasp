// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
)

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

func AddSilentNodes[Obj any](tc *TestContext[Obj], silentNodes []NodeID) {
	nodes := tc.Nodes()
	for _, nid := range silentNodes {
		if _, exists := nodes[nid]; exists {
			panic(fmt.Errorf("node %s already exists in the test context - cannot add as silent node", nid.ShortString()))
		}
		nodes[nid] = lo.Empty[Obj]()
	}
	tc.SetNodes(nodes)

	functors := tc.Functors()

	origApplyInput := functors.ApplyInput
	functors.ApplyInput = func(nid NodeID, obj Obj, input Input) []MessageOut {
		if lo.Contains(silentNodes, nid) {
			return nil
		}
		return origApplyInput(nid, obj, input)
	}

	origApplyMessage := functors.ApplyMessage
	functors.ApplyMessage = func(nid NodeID, obj Obj, msg MessageIn[any]) []MessageOut {
		if lo.Contains(silentNodes, nid) {
			return nil
		}
		return origApplyMessage(nid, obj, msg)
	}

	origOutput := functors.Output
	functors.Output = func(nid NodeID, obj Obj) any {
		if lo.Contains(silentNodes, nid) {
			return nil
		}
		return origOutput(nid, obj)
	}

	origStatusString := functors.StatusString
	functors.StatusString = func(nid NodeID, obj Obj) string {
		if lo.Contains(silentNodes, nid) {
			return fmt.Sprintf("StatusString{silentNode=%v}", nid)
		}
		return origStatusString(nid, obj)
	}

	tc.SetFunctors(functors)
}
