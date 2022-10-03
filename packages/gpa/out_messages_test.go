// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/wasp/packages/gpa"
)

func TestOutMessages(t *testing.T) {
	// Create a fresh one.
	msgs := gpa.NoMessages()
	assert.Equal(t, 0, msgs.Count())
	assert.Equal(t, []gpa.Message{}, msgs.AsArray())
	//
	// Add some messages.
	m1 := &gpa.TestMessage{ID: 1}
	m2 := &gpa.TestMessage{ID: 2}
	m3 := &gpa.TestMessage{ID: 3}
	m4 := &gpa.TestMessage{ID: 4}
	m5 := &gpa.TestMessage{ID: 5}
	msgs.Add(m1)
	msgs.AddMany([]gpa.Message{m2, m3, m4})
	msgs.Add(m5)
	assert.Equal(t, 5, msgs.Count())
	assert.Equal(t, []gpa.Message{m1, m2, m3, m4, m5}, msgs.AsArray())
	//
	// Add one to other.
	m0 := &gpa.TestMessage{ID: 0}
	moreMsgs := gpa.NoMessages().Add(m0).AddAll(msgs)
	assert.Equal(t, 6, moreMsgs.Count())
	assert.Equal(t, []gpa.Message{m0, m1, m2, m3, m4, m5}, moreMsgs.AsArray())
}

// Check if appending works while iterating.
func TestOutMessagesIterate(t *testing.T) {
	m1 := &gpa.TestMessage{ID: 1}
	m2 := &gpa.TestMessage{ID: 2}
	m3 := &gpa.TestMessage{ID: 3}
	m4 := &gpa.TestMessage{ID: 4}

	out := []gpa.Message{}
	msgs := gpa.NoMessages().Add(m1)
	msgs.MustIterate(func(msg gpa.Message) {
		out = append(out, msg)
		if msg.(*gpa.TestMessage).ID == 1 {
			msgs.AddMany([]gpa.Message{m2, m3, m4})
		}
	})
	assert.Equal(t, 4, msgs.Count())
	assert.Equal(t, []gpa.Message{m1, m2, m3, m4}, msgs.AsArray())
}
