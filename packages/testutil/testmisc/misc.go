package testmisc

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/stretchr/testify/assert"
)

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
