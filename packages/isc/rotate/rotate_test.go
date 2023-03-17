package rotate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestBasicRotateRequest(t *testing.T) {
	kp, addr := testkey.GenKeyAddr()
	req := NewRotateRequestOffLedger(isc.RandomChainID(), addr, kp, gas.LimitsDefault.MaxGasPerRequest)
	require.True(t, IsRotateStateControllerRequest(req))
}
