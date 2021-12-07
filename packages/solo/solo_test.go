package solo

import "testing"

func TestSoloBasic(t *testing.T) {
	env := New(t, false, false)
	_ = env.NewChain(nil, "ch1")
}
