// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

type inputCanPropose struct{}

func NewInputCanPropose() *inputCanPropose {
	return &inputCanPropose{}
}

func (inp *inputCanPropose) String() string {
	return "{chainMgr.inputCanPropose}"
}
