package testmisc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func RequireErrorToBe(t *testing.T, err error, target interface{}) {
	if err == nil {
		t.Errorf("error expected, found nil")
		t.FailNow()
		return
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
	t.Errorf("error does not contain '%s' but instead is '%v'", targ, err)
	t.FailNow()
}
