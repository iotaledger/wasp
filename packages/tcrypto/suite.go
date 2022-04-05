// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/key"
)

type Suite interface {
	kyber.Group
	pairing.Suite
	key.Suite
}

func DefaultEd25519Suite() suites.Suite {
	return suites.MustFind("Ed25519")
}

func DefaultBlsSuite() Suite {
	return pairing.NewSuiteBn256()
}
