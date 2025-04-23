package providers

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
)

type MnemonicSeed struct {
	cryptolib.Signer
	addressIndex uint32
}

func NewMnemonicSeed(keyPair *cryptolib.KeyPair) *MnemonicSeed {
	return &MnemonicSeed{
		Signer:       keyPair,
		addressIndex: 0,
	}
}

func (i *MnemonicSeed) AddressIndex() uint32 {
	// Derivation unsupported
	return 0
}

func LoadMnemonicSeed() wallets.Wallet {
	s, _ := iotasigner.KeyFromMnemonic(config.GetTestingMnemonic(), iotasigner.KeySchemeFlagDefault)
	rawSeed := s.RawSeed()
	kp := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes(rawSeed[:]))
	return NewMnemonicSeed(kp)
}
