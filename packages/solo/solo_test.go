package solo

import "testing"

func TestSoloBasic(t *testing.T) {
	env := New(t, true, false)
	_ = env.NewChain(nil, "ch1")
}
