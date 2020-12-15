// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package dkg is responsible for performing a distributed key
// generation procedure. The client-side part is implemented
// as an initiator, and the nodes sharing a generated secret
// are implemented as Node.
//
// Implementation is based on <https://github.com/dedis/kyber/blob/master/share/dkg/rabin/dkg.go>
// which is based on <https://link.springer.com/article/10.1007/s00145-006-0347-3>.
package dkg

// TODO: Only authenticated nodes can initiate (and participate in?) the DKG.
