// only included if runheavy tag is not present
//go:build !runheavy

package testutil

import "testing"

//nolint:gocritic // its not a test function, but gets called by other test functions
func RunHeavy(t *testing.T) {
	t.Logf("skipping heavy test %s", t.Name())
	t.SkipNow()
}
