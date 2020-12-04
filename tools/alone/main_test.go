package alone

import (
	"testing"
)

func TestBasic(t *testing.T) {
	InitEnvironment(t)
	//t.Logf("\n%s", env.String())
	Env.Infof("\n%s\n", Env.String())
}
