package coreutil

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/wasp/packages/testutil/testkey"

	"github.com/stretchr/testify/require"
)

func TestBasicRotateRequest(t *testing.T) {
	kp, addr := testkey.GenKeyAddr()
	req := NewRotateRequestOffLedger(addr, kp)
	_, _ = req.SolidifyArgs(coretypes.NewInMemoryBlobCache())
	require.True(t, IsRotateCommitteeRequest(req))
}
