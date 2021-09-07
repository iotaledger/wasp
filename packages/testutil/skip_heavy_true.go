// only included if tag=skipheavy
//+build skipheavy

package testutil

import "testing"

func SkipHeavy(t *testing.T) {
	t.Logf("skipping heavy test %s", t.Name())
	t.SkipNow()
}
