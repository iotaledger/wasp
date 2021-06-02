package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

// NewSignatureSchemeWithFundsAndPubKey generates new ed25519 signature scheme
// and requests some tokens from the UTXODB faucet.
// The amount of tokens is equal to solo.Saldo (=1000000) iotas
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPairWithFunds() (*ed25519.KeyPair, ledgerstate.Address) {
	keyPair, addr := env.NewKeyPair()

	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	_, err := env.utxoDB.RequestFunds(addr, env.LogicalTime())
	require.NoError(env.T, err)
	env.AssertAddressBalance(addr, ledgerstate.ColorIOTA, Saldo)

	return keyPair, addr
}

// NewSignatureSchemeAndPubKey generates new ed25519 signature scheme
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPair() (*ed25519.KeyPair, ledgerstate.Address) {
	keyPair := ed25519.GenerateKeyPair()
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	env.AssertAddressBalance(addr, ledgerstate.ColorIOTA, 0)
	return &keyPair, ledgerstate.NewED25519Address(keyPair.PublicKey)
}

// MintTokens mints specified amount of new colored tokens in the given wallet (signature scheme)
// Returns the color of minted tokens: the hash of the transaction
func (env *Solo) MintTokens(wallet *ed25519.KeyPair, amount uint64) (ledgerstate.Color, error) {
	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	addr := ledgerstate.NewED25519Address(wallet.PublicKey)
	allOuts := env.utxoDB.GetAddressOutputs(addr)

	txb := utxoutil.NewBuilder(allOuts...).WithTimestamp(env.LogicalTime())
	if amount < DustThresholdIotas {
		return [32]byte{}, xerrors.New("can't mint number of tokens below dust threshold")
	}
	if err := txb.AddMintingOutputConsume(addr, amount); err != nil {
		return [32]byte{}, err
	}
	if err := txb.AddRemainderOutputIfNeeded(addr, nil, true); err != nil {
		return [32]byte{}, err
	}
	tx, err := txb.BuildWithED25519(wallet)
	if err != nil {
		return [32]byte{}, err
	}
	if err := env.AddToLedger(tx); err != nil {
		return [32]byte{}, nil
	}
	m := utxoutil.GetMintedAmounts(tx)
	require.EqualValues(env.T, 1, len(m))

	var ret ledgerstate.Color
	for col := range m {
		ret = col
		break
	}
	return ret, nil
}

func (env *Solo) PutBlobDataIntoRegistry(data []byte) hashing.HashValue {
	h, err := env.blobCache.PutBlob(data)
	require.NoError(env.T, err)
	env.logger.Infof("Solo::PutBlobDataIntoRegistry: len = %d, hash = %s", len(data), h)
	return h
}
