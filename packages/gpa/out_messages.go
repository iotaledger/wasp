// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

type outMessagesElem struct {
	msg  Message     //  / One
	msgs []Message   // {  of these
	out  OutMessages //  \ can be present.
	next *outMessagesElem
}

type outMessagesImpl struct {
	count int
	done  bool
	head  *outMessagesElem
	tail  *outMessagesElem
}

var _ OutMessages = &outMessagesImpl{}

// A convenience function to return from the Input or Message functions in GPA.
func NoMessages() OutMessages {
	return &outMessagesImpl{count: 0, head: nil, tail: nil}
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) Add(msg Message) OutMessages {
	if msg == nil {
		panic("trying to add nil message, is that a mistake?")
	}
	if omi.done {
		panic("out messages marked as done")
	}
	omi.addElem(&outMessagesElem{msg: msg, next: nil})
	omi.count++
	return omi
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) AddMany(msgs []Message) OutMessages {
	if omi.done {
		panic("out messages marked as done")
	}
	if msgs == nil || len(msgs) == 0 {
		return omi
	}
	omi.addElem(&outMessagesElem{msgs: msgs, next: nil})
	omi.count += len(msgs)
	return omi
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) AddAll(msgs OutMessages) OutMessages {
	if omi.done {
		panic("out messages marked as done")
	}
	if omi == msgs {
		panic("cannot append self to itself")
	}
	if msgs == nil || msgs.Count() == 0 {
		return omi
	}
	omi.addElem(&outMessagesElem{out: msgs.Done(), next: nil})
	omi.count += msgs.Count()
	return omi
}

func (omi *outMessagesImpl) addElem(elem *outMessagesElem) {
	if omi.head == nil {
		omi.head = elem
		omi.tail = omi.head
		return
	}
	omi.tail.next = elem
	omi.tail = elem
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) Done() OutMessages {
	omi.done = true
	return omi
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) Count() int {
	return omi.count
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) Iterate(callback func(msg Message) error) error {
	for elem := omi.head; elem != nil; elem = elem.next {
		if elem.msg != nil {
			if err := callback(elem.msg); err != nil {
				return err
			}
			continue
		}
		if elem.msgs != nil {
			for i := range elem.msgs {
				if err := callback(elem.msgs[i]); err != nil {
					return err
				}
			}
			continue
		}
		if err := elem.out.Iterate(callback); err != nil {
			return err
		}
	}
	return nil
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) MustIterate(callback func(msg Message)) {
	for elem := omi.head; elem != nil; elem = elem.next {
		if elem.msg != nil {
			callback(elem.msg)
			continue
		}
		if elem.msgs != nil {
			for i := range elem.msgs {
				callback(elem.msgs[i])
			}
			continue
		}
		elem.out.MustIterate(callback)
	}
}

// Implements the OutMessages interface.
func (omi *outMessagesImpl) AsArray() []Message {
	out := make([]Message, omi.count)
	pos := 0
	omi.MustIterate(func(msg Message) {
		out[pos] = msg
		pos++
	})
	return out
}
