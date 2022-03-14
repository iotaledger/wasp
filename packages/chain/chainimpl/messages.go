// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

// DismissChainMsg sent by component to the chain core in case of major setback
type DismissChainMsg struct {
	Reason string
}
