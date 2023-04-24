// only included if tag=runheavy
//go:build runheavy
// +build runheavy

package testutil

import "testing"

//nolint:gocritic // its not a test function, but gets called by other test functions
func RunHeavy(t *testing.T) {
}
