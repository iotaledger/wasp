// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

// This is done just to decrease number of diff lines in PR. We can change it after demo of idea.
type OutMessages *OutMessagesV
type OutMessagesV struct {
	NextLogIndex []MsgNextLogIndex
}

func (m *OutMessagesV) AddAll(msgs OutMessages) OutMessages {
	if msgs != nil {
		m.NextLogIndex = append(m.NextLogIndex, msgs.NextLogIndex...)
	}
	return m
}

func NoMessages() *OutMessagesV {
	return &OutMessagesV{}
}
