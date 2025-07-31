package solo

import (
	"hash/fnv"
	"math"
	"time"

	"fortio.org/safecast"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
)

func (env *Solo) L1Client() clients.L1Client {
	return clients.NewL1Client(clients.L1Config{
		APIURL:    env.l1Config.IotaRPCURL,
		FaucetURL: env.l1Config.IotaFaucetURL,
	}, iotaclient.WaitForEffectsEnabled)
}

func (env *Solo) ISCMoveClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		env.l1Config.IotaRPCURL,
		env.l1Config.IotaFaucetURL,
		l1starter.WaitUntilEffectsVisible,
	)
}

func (env *Solo) NewKeyPairFromIndex(index int) *cryptolib.KeyPair {
	var seed cryptolib.Seed
	if testkey.UseRandomSeed() {
		seed = cryptolib.NewSeed()
	} else {
		seed = *env.NewSeedFromIndex(index)
	}
	return cryptolib.KeyPairFromSeed(seed)
}

func (env *Solo) NewSeedFromIndex(index int) *cryptolib.Seed {
	if index < 0 {
		// SubSeed takes a "uint31"
		index += math.MaxUint32 / 2
	}
	seed := cryptolib.SubSeed(env.seed[:], safecast.MustConvert[uint32](index))
	return &seed
}

func (env *Solo) NewSeedFromTestNameAndTimestamp(testName string) *cryptolib.Seed {
	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(testName + time.Now().String()))
	if err != nil {
		panic(err)
	}

	seed := cryptolib.SubSeed(env.seed[:], algorithm.Sum32()/2)
	return &seed
}

func (env *Solo) WaitForNewBalance(address *cryptolib.Address, startBalance coin.Value) {
	const amountOfRetries = 10
	currentBalance := env.L1BaseTokens(address)

	if currentBalance > startBalance {
		return
	}

	count := 0
	for {
		if count == amountOfRetries {
			panic("Could not get funds from Faucet")
		}

		if env.L1BaseTokens(address) > startBalance {
			break
		}

		count++
		time.Sleep(1 * time.Second)
	}
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
	currentBalance := env.L1BaseTokens(target)
	err := iotaclient.RequestFundsFromFaucet(env.ctx, target.AsIotaAddress(), env.l1Config.IotaFaucetURL)
	env.WaitForNewBalance(target, currentBalance)
	require.NoError(env.T, err)
	require.GreaterOrEqual(env.T, env.L1BaseTokens(target), coin.Value(iotaclient.FundsFromFaucetAmount))
}

func (env *Solo) NewKeyPair(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, *cryptolib.Address) {
	return testkey.GenKeyAddr(seedOpt...)
}
