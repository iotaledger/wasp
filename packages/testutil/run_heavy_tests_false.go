// only included if runheavy tag is not present
//go:build !runheavy
// +build !runheavy

package testutil

import "testing"

func RunHeavy(t *testing.T) {
	t.Logf("skipping heavy test %s", t.Name())
	t.SkipNow()
}
