package suisigner

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/sui-go/sui"
)

const (
	SignatureFlagEd25519   = 0x0
	SignatureFlagSecp256k1 = 0x1

	// IOTA_DIFF 4218 is for iota
	DerivationPathEd25519   = `m/44'/4218'/0'/0'/0'`
	DerivationPathSecp256k1 = `m/54'/4218'/0'/0/0`
)

type Signer interface {
	Address() *sui.Address
	Sign(msg []byte) (signature *Signature, err error)
	SignTransactionBlock(txnBytes []byte, intent Intent) (*Signature, error)
}

// FIXME support more than ed25519
type InMemorySigner struct {
	ed25519Keypair *KeypairEd25519
	// secp256k1Keypair *KeypairSecp256k1

	address *sui.Address
}

func NewSigner(seed []byte, flag KeySchemeFlag) *InMemorySigner {
	prikey := ed25519.NewKeyFromSeed(seed[:])
	pubkey := prikey.Public().(ed25519.PublicKey)

	// IOTA_DIFF iota ignore flag when signature scheme is ed25519
	var buf []byte
	switch flag {
	case KeySchemeFlagEd25519:
		buf = []byte{KeySchemeFlagEd25519.Byte()}
	case KeySchemeFlagSecp256k1:
		buf = []byte{KeySchemeFlagEd25519.Byte()}
	case KeySchemeFlagIotaEd25519:
		buf = []byte{}
	default:
		panic("unrecognizable key scheme flag")
	}
	buf = append(buf, pubkey...)
	addrBytes := blake2b.Sum256(buf)
	addr := "0x" + hex.EncodeToString(addrBytes[:])

	return &InMemorySigner{
		ed25519Keypair: &KeypairEd25519{
			PriKey: prikey,
			PubKey: pubkey,
		},
		address: sui.MustAddressFromHex(addr),
	}
}

// there are only 256 different signers can be generated
func NewSignerByIndex(seed []byte, flag KeySchemeFlag, index int) Signer {
	seed[0] = seed[0] + byte(index)
	return NewSigner(seed, flag)
}

// generate keypair (signer) with mnemonic which is referring the Sui monorepo in the following code
//
// let phrase = "asset pink record dawn hundred sure various crime client enforce carbon blossom";
// let mut keystore = Keystore::from(InMemKeystore::new_insecure_for_tests(0));
// let generated_address = keystore.import_from_mnemonic(&phrase, SignatureScheme::ED25519, None, None).unwrap();
func NewSignerWithMnemonic(mnemonic string, flag KeySchemeFlag) (Signer, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}
	key, err := DeriveForPath(DerivationPathEd25519, seed)
	if err != nil {
		return nil, err
	}
	return NewSigner(key.Key, flag), nil
}

func (s *InMemorySigner) PrivateKey() []byte {
	switch {
	case s.ed25519Keypair != nil:
		return s.ed25519Keypair.PriKey
	default:
		return nil
	}
}

func (s *InMemorySigner) PublicKey() []byte {
	switch {
	case s.ed25519Keypair != nil:
		return s.ed25519Keypair.PubKey
	default:
		return nil
	}
}

func (s *InMemorySigner) Sign(msg []byte) (signature *Signature, err error) {
	// FIXME support more than ed25519
	sig := ed25519.Sign(s.ed25519Keypair.PriKey, msg)

	return &Signature{
		Ed25519Signature: NewEd25519Signature(s.ed25519Keypair.PubKey, sig),
	}, nil
}

func (a *InMemorySigner) Address() *sui.Address {
	return a.address
}

// FIXME support more than ed25519
func (a *InMemorySigner) SignTransactionBlock(txnBytes []byte, intent Intent) (*Signature, error) {
	data := MessageWithIntent(intent, bcsBytes(txnBytes))
	hash := blake2b.Sum256(data)
	return a.Sign(hash[:])
}

type bcsBytes []byte

func (b bcsBytes) MarshalBCS() ([]byte, error) {
	return b, nil
}
