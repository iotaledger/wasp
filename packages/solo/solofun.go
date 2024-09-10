package solo

import (
	"math"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/sui-go/suiclient"
)

func (env *Solo) SuiClient() *suiclient.Client {
	return suiclient.NewHTTP(env.l1Config.SuiRPCURL)
}

func (env *Solo) ISCMoveClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		env.l1Config.SuiRPCURL,
		env.l1Config.SuiFaucetURL,
	)
}

func (env *Solo) NewKeyPairFromIndex(index int) *cryptolib.KeyPair {
	seed := env.NewSeedFromIndex(index)
	return cryptolib.KeyPairFromSeed(*seed)
}

func (env *Solo) NewSeedFromIndex(index int) *cryptolib.Seed {
	if index < 0 {
		// SubSeed takes a "uint31"
		index += math.MaxUint32 / 2
	}
	seed := cryptolib.SubSeed(env.seed[:], uint32(index))
	return &seed
}

// NewKeyPairWithFunds generates new ed25519 signature scheme
// and requests some tokens from the UTXODB faucet.
// The amount of tokens is equal to utxodb.FundsFromFaucetAmount (=1000Mi) base tokens
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPairWithFunds(seed ...*cryptolib.Seed) (*cryptolib.KeyPair, *cryptolib.Address) {
	keyPair, addr := env.NewKeyPair(seed...)
	env.GetFundsFromFaucet(addr)
	return keyPair, addr
}

func (env *Solo) GetFundsFromFaucet(target *cryptolib.Address) {
	err := suiclient.RequestFundsFromFaucet(env.ctx, target.AsSuiAddress(), env.l1Config.SuiFaucetURL)
	require.NoError(env.T, err)
	env.AssertL1BaseTokens(target, suiclient.FundsFromFaucetAmount)
}

// NewSignatureSchemeAndPubKey generates new ed25519 signature scheme
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPair(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, *cryptolib.Address) {
	return testkey.GenKeyAddr(seedOpt...)
}
