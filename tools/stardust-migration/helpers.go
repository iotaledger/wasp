package main

import (
	"fmt"

	"github.com/nnikolash/wasp-types-exported/packages/kv"
)

// Wraps a function and adds to it printing of number of times it was called
func p[Src any, Dest any](f EntityMigrationFunc[Src, Dest]) EntityMigrationFunc[Src, Dest] {
	callCount := 0

	return func(oldKey kv.Key, srcVal Src) (kv.Key, Dest) {
		callCount++
		if callCount%100 == 0 {
			fmt.Printf("\rProcessed: %v         ", callCount)
		}

		return f(oldKey, srcVal)
	}
}
