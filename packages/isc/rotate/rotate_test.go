package rotate

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/stretchr/testify/require"
)

func TestBasicRotateRequest(t *testing.T) {
	kp, addr := testkey.GenKeyAddr()
	req := NewRotateRequestOffLedger(isc.RandomChainID(), addr, kp)
	require.True(t, IsRotateStateControllerRequest(req))
}
