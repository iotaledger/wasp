package tests

import (
	"testing"

	"github.com/iotaledger/wasp/packages/testutil"
)

func TestSkipHeavy(t *testing.T) {
	testutil.SkipHeavy(t)
	t.Logf("running the test")
}
