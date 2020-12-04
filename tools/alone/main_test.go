package alone

import (
	"testing"
)

func TestBasic(t *testing.T) {
	InitEnvironment(t)
	Env.Infof("\n%s\n", Env.String())
}
