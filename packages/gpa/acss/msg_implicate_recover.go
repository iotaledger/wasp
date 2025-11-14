// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acss

type msgImplicateKind byte

const (
	msgImplicateRecoverKindIMPLICATE msgImplicateKind = iota
	msgImplicateRecoverKindRECOVER
)

// The <IMPLICATE, i, skᵢ> and <RECOVER, i, skᵢ> messages.
type MsgImplicateRecover struct {
	kind msgImplicateKind `bcs:"export"`
	i    int              `bcs:"export,type=u16"`
	data []byte           `bcs:"export"` // Either implication or the recovered secret.
}
