// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

const msgTypeTest MessageType = 0xff

// TestMessage is just a message for test cases.
type TestMessage struct {
	ID int
}
