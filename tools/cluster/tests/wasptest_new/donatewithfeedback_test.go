package wasptest

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

const dwfName = "donatewithfeedback"

const dwfFile = "wasm/donatewithfeedback_bg.wasm"

const dwfDescription = "Donate with feedback, a PoC smart contract"

var dwfHname = coretypes.Hn(dwfName)

func TestDwfDonateOnce(t *testing.T) {
	const numDonations = 1
	al := solo.New(t, false, false)
	chain := al.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, dwfName, dwfFile)
	require.NoError(t, err)

	for i := 0; i < numDonations; i++ {
		feedback := fmt.Sprintf("Donation #%d: well done, I give you 42 iotas", i)
		req := solo.NewCallParams(dwfName, "donate", "feedback", feedback).
			WithTransfer(balance.ColorIOTA, 42)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	ret, err := chain.CallView(dwfName, "donations")
	require.NoError(t, err)
	largest, _, err := codec.DecodeInt64(ret.MustGet("maxDonation"))
	check(err, t)
	require.EqualValues(t, 42, largest)
	total, _, err := codec.DecodeInt64(ret.MustGet("totalDonation"))
	check(err, t)
	require.EqualValues(t, 42*numDonations, total)
}
