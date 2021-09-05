package testutil

import "testing"

// add 'tags=skipheavy' to the go test command line and tests which invoke SkipHeavy(t) will be skipped
var skipHeavy = false

func SkipHeavy(t *testing.T) {
	if skipHeavy {
		t.Logf("skipping heavy test %s", t.Name())
		t.SkipNow()
	}
}
