package solo

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/stretchr/testify/require"
)

// NewSignatureSchemeWithFunds generates new ed25519 signature scheme
// and requests some tokens from the UTXODB faucet.
// The amount of tokens is equal to solo.Saldo (=1337) iotas
func (env *Solo) NewSignatureSchemeWithFunds() signaturescheme.SignatureScheme {
	ret, _ := env.NewSignatureSchemeWithFundsAndPubKey()
	return ret
}

// NewSignatureSchemeWithFundsAndPubKey generates new ed25519 signature scheme
// and requests some tokens from the UTXODB faucet.
// The amount of tokens is equal to solo.Saldo (=1337) iotas
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewSignatureSchemeWithFundsAndPubKey() (signaturescheme.SignatureScheme, []byte) {
	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	ret, pubKeyBytes := env.NewSignatureSchemeAndPubKey()
	_, err := env.utxoDB.RequestFunds(ret.Address())
	require.NoError(env.T, err)
	return ret, pubKeyBytes
}

// NewSignatureScheme generates new ed25519 signature scheme
func (env *Solo) NewSignatureScheme() signaturescheme.SignatureScheme {
	ret, _ := env.NewSignatureSchemeAndPubKey()
	return ret
}

// NewSignatureSchemeAndPubKey generates new ed25519 signature scheme
// Returns signature scheme interface and public key in binary form
func (env *Solo) NewSignatureSchemeAndPubKey() (signaturescheme.SignatureScheme, []byte) {
	keypair := ed25519.GenerateKeyPair()
	ret := signaturescheme.ED25519(keypair)
	env.AssertAddressBalance(ret.Address(), balance.ColorIOTA, 0)
	return ret, keypair.PublicKey.Bytes()
}

// MintTokens mints specified amount of new colored tokens in the given wallet (signature scheme)
// Returns the color of minted tokens: the hash of the transaction
func (env *Solo) MintTokens(wallet signaturescheme.SignatureScheme, amount int64) (balance.Color, error) {
	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	allOuts := env.utxoDB.GetAddressOutputs(wallet.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(env.T, err)

	if err = txb.MintColoredTokens(wallet.Address(), balance.ColorIOTA, amount); err != nil {
		return balance.Color{}, err
	}
	tx := txb.BuildValueTransactionOnly(false)
	tx.Sign(wallet)

	if err = env.utxoDB.AddTransaction(tx); err != nil {
		return balance.Color{}, err
	}
	return balance.Color(tx.ID()), nil
}

// DestroyColoredTokens uncolors specified amount of colored tokens, i.e. converts them into IOTAs
func (env *Solo) DestroyColoredTokens(wallet signaturescheme.SignatureScheme, color balance.Color, amount int64) error {
	env.ledgerMutex.Lock()
	defer env.ledgerMutex.Unlock()

	allOuts := env.utxoDB.GetAddressOutputs(wallet.Address())
	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	require.NoError(env.T, err)

	if err = txb.EraseColor(wallet.Address(), color, amount); err != nil {
		return err
	}
	tx := txb.BuildValueTransactionOnly(false)
	tx.Sign(wallet)

	return env.utxoDB.AddTransaction(tx)
}

func (env *Solo) PutBlobDataIntoRegistry(data []byte) hashing.HashValue {
	h, err := env.registry.PutBlob(data)
	require.NoError(env.T, err)
	env.logger.Infof("Solo::PutBlobDataIntoRegistry: len = %d, hash = %s", len(data), h)
	return h
}
