package vm0poc

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
)

func TestBasic(t *testing.T) {
	env := solo.New(t)
	_ = env.NewChain(nil, "ch1", solo.InitChainOptions{VMRunner: NewVMRunner()})
}
