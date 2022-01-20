package solo

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/stretchr/testify/require"
)

func (env *Solo) NewSeedFromIndex(index int) *cryptolib.Seed {
	seed := cryptolib.SeedFromByteArray(hashing.HashData(env.utxoDB.Seed(), util.Int32To4Bytes(int32(index))).Bytes())
	return &seed
}

// NewSignatureSchemeWithFundsAndPubKey generates new ed25519 signature scheme
// and requests some tokens from the UTXODB faucet.
// The amount of tokens is equal to solo.Saldo (=1000000) iotas
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPairWithFunds(seed ...*cryptolib.Seed) (*cryptolib.KeyPair, iotago.Address) {
	keyPair, addr := env.NewKeyPair(seed...)

	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	_, err := env.utxoDB.GetFundsFromFaucet(addr)
	require.NoError(env.T, err)
	env.AssertL1Iotas(addr, Saldo)

	return keyPair, addr
}

func (env *Solo) GetFundsFromFaucet(target iotago.Address, amount ...uint64) (*iotago.Transaction, error) {
	return env.utxoDB.GetFundsFromFaucet(target, amount...)
}

// NewSignatureSchemeAndPubKey generates new ed25519 signature scheme
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPair(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, iotago.Address) {
	return testkey.GenKeyAddr(seedOpt...)
}

// MintTokens mints specified amount of new colored tokens in the given wallet (signature scheme)
// Returns the color of minted tokens: the hash of the transaction
func (env *Solo) MintTokens(wallet *cryptolib.KeyPair, amount uint64) (iotago.NativeTokenID, error) {
	panic("not implemented")
	// env.ledgerMutex.Lock()
	// defer env.ledgerMutex.Unlock()

	// addr := ledgerstate.NewED25519Address(wallet.PublicKey)
	// allOuts := env.utxoDB.GetAddressOutputs(addr)

	// txb := utxoutil.NewBuilder(allOuts...).WithTimestamp(env.GlobalTime())
	// if amount < DustThresholdIotas {
	// 	return colored.Color{}, xerrors.New("can't mint number of tokens below dust threshold")
	// }
	// if err := txb.AddMintingOutputConsume(addr, amount); err != nil {
	// 	return colored.Color{}, err
	// }
	// if err := txb.AddRemainderOutputIfNeeded(addr, nil, true); err != nil {
	// 	return colored.Color{}, err
	// }
	// tx, err := txb.BuildWithED25519(wallet)
	// if err != nil {
	// 	return colored.Color{}, err
	// }
	// if err := env.AddToLedger(tx); err != nil {
	// 	return colored.Color{}, nil
	// }
	// m := utxoutil.GetMintedAmounts(tx)
	// require.EqualValues(env.T, 1, len(m))

	// var ret colored.Color
	// for col := range m {
	// 	ret = colored.ColorFromL1Color(col)
	// 	break
	// }
	// return ret, nil
}
