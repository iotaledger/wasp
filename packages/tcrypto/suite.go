// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package tcrypto

import (
	"sync"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/kyber/v3/util/key"
)

var (
	ed25519Suite     suites.Suite
	ed25519SuiteOnce sync.Once

	blsSuite     Suite
	blsSuiteOnce sync.Once
)

type Suite interface {
	kyber.Group
	pairing.Suite
	key.Suite
}

func DefaultEd25519Suite() suites.Suite {
	ed25519SuiteOnce.Do(func() {
		if ed25519Suite == nil {
			ed25519Suite = suites.MustFind("Ed25519")
		}
	})

	return ed25519Suite
}

func DefaultBLSSuite() Suite {
	blsSuiteOnce.Do(func() {
		if blsSuite == nil {
			blsSuite = pairing.NewSuiteBn256()
		}
	})

	return blsSuite
}
