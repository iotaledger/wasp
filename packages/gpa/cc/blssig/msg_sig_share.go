// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package blssig

type MsgSigShare struct {
	sigShare []byte `bcs:"export"`
}
