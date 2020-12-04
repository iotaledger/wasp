package alone

import (
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasic(t *testing.T) {
	InitEnvironment(t)
	Env.Infof("\n%s\n", Env.String())
}

func TestPost(t *testing.T) {
	InitEnvironment(t)

	req := NewRequest(Env.OriginatorSigscheme, root.Interface.Name, root.FuncGetInfo)
	res, err := Env.PostRequest(req)
	require.NoError(t, err)

	Env.Infof("result from getInfo: %s", res.String())
}
