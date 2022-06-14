// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
)

type SecretShare struct {
	share   *share.PriShare
	commits []kyber.Point
}

func NewSecretShare(priShare *share.PriShare, commits []kyber.Point) *SecretShare {
	return &SecretShare{share: priShare, commits: commits}
}

// PriShare returns the private share.
func (s *SecretShare) PriShare() *share.PriShare {
	return s.share
}

// Commitments returns the coefficients of the public polynomial.
func (s *SecretShare) Commitments() []kyber.Point {
	return s.commits
}
