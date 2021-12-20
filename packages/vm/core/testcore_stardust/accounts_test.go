package testcore

import (
	"math/big"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/stretchr/testify/require"
)

func TestCreateFoundry(t *testing.T) {
	env := solo.New(t)
	env.EnablePublisher(true)
	ch := env.NewChain(nil, "chain1")
	defer ch.Log.Sync()

	senderKeyPair, senderAddr := env.NewKeyPairWithFunds(env.NewSeedFromIndex(10))
	_ = iscp.NewAgentID(senderAddr, 0)

	req := solo.NewCallParams(accounts.Contract.Name, accounts.FuncFoundryCreateNew.Name,
		accounts.ParamsTokenTag, codec.EncodeTokenTag(iotago.TokenTag{}),
		accounts.ParamsMaxSupply, big.NewInt(10).Bytes(),
	)
	_, _, err := ch.PostRequestSyncTx(req, senderKeyPair)
	require.NoError(t, err)
	env.WaitPublisher()
}
