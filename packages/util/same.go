// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import "golang.org/x/exp/slices"

type Equated[V any] interface {
	Equals(other V) bool
}

// Checks two slices, if they represent the same sets (same elements, maybe a different order).
func Same[V Equated[V]](a, b []V) bool {
	if len(a) != len(b) {
		return false
	}
	for _, e := range a {
		if slices.IndexFunc(b, e.Equals) == -1 {
			return false
		}
	}
	return true
}
