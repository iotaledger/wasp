// only included if skipheavy tag is not present
//go:build !skipheavy
// +build !skipheavy

package testutil

import "testing"

func SkipHeavy(t *testing.T) {
}
