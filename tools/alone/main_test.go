package alone

import (
	"testing"
)

func TestBasic(t *testing.T) {
	env := NewEnvironment(t)
	//t.Logf("\n%s", env.String())
	env.Infof("\n%s\n", env.String())
}
