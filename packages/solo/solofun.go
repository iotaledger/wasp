package solo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/stretchr/testify/require"
)

// NewSignatureSchemeWithFundsAndPubKey generates new ed25519 signature scheme
// and requests some tokens from the UTXODB faucet.
// The amount of tokens is equal to solo.Saldo (=1337) iotas
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewKeyPairWithFunds() (*ed25519.KeyPair, ledgerstate.Address) {
	keyPair, addr := env.NewKeyPair()

	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	_, err := env.utxoDB.RequestFunds(addr)
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

	txb := utxoutil.NewBuilder(allOuts...)
	numIotas := DustThresholdIotas
	if amount > numIotas {
		numIotas = amount
	}
	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: amount}

	if err := txb.AddExtendedOutputSimple(addr, nil, bals, amount); err != nil {
		return [32]byte{}, err
	}
	if err := txb.AddReminderOutputIfNeeded(addr, nil, true); err != nil {
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

// DestroyColoredTokens uncolors specified amount of colored tokens, i.e. converts them into IOTAs
//func (env *Solo) DestroyColoredTokens(wallet *ed25519.KeyPair, color ledgerstate.Color, amount uint64) error {
//	env.ledgerMutex.Lock()
//	defer env.ledgerMutex.Unlock()
//
//	addr := ledgerstate.NewED25519Address(wallet.PublicKey)
//
//	allOuts := env.utxoDB.GetAddressOutputs(addr)
//	utxoutil.ConsumeRemaining(allOuts...)
//	if !utxoutil.ConsumeColored(color, amount){
//		return xerrors.New("not enough balance")
//	}
//
//	txb := utxoutil.NewBuilder(allOuts...)
//	numIotas := DustThresholdIotas
//	if amount > numIotas {
//		numIotas = amount
//	}
//	bals := map[ledgerstate.Color]uint64{ledgerstate.}
//
//	allOuts := env.utxoDB.GetAddressOutputs(wallet.Address())
//	txb, err := txbuilder.NewFromOutputBalances(allOuts)
//	require.NoError(env.T, err)
//
//	if err = txb.EraseColor(wallet.Address(), color, amount); err != nil {
//		return err
//	}
//	tx := txb.BuildValueTransactionOnly(false)
//	tx.Sign(wallet)
//
//	return env.utxoDB.AddTransaction(tx)
//}

func (env *Solo) PutBlobDataIntoRegistry(data []byte) hashing.HashValue {
	h, err := env.blobCache.PutBlob(data)
	require.NoError(env.T, err)
	env.logger.Infof("Solo::PutBlobDataIntoRegistry: len = %d, hash = %s", len(data), h)
	return h
}
