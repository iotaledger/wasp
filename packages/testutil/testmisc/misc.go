// Package testmisc provides miscellaneous utility functions for testing
package testmisc

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

// GetTimeout returns a Duration that is three times longer than the localTimeout if the tests are being run via GitHub
// Actions.
func GetTimeout(localTimeout time.Duration) time.Duration {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		localTimeout *= 3
	}
	return localTimeout
}

func RequireErrorToBe(t *testing.T, err error, target interface{}) {
	if err == nil {
		assert.Fail(t, "error expected, found nil")
		t.Fatal()
		return
	}
	if target, ok := target.(isc.VMErrorBase); ok {
		if isc.VMErrorIs(err, target) {
			return
		}
	}
	var targ string
	switch target := target.(type) {
	case error:
		if errors.Is(err, target) {
			return
		}
		targ = target.Error()
	case string:
		targ = target
	case interface{ String() string }:
		targ = target.String()
	default:
		panic(fmt.Sprintf("RequireErrorToBe: type %T not supported", target))
	}
	if strings.Contains(err.Error(), targ) {
		return
	}
	assert.Fail(t, fmt.Sprintf("error does not contain '%s' but instead is '%v'", targ, err))
	t.Fatal()
}
