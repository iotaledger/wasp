package wallets

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type ExternalAddressSigner struct {
	keyPair cryptolib.VariantKeyPair
}

func (r *ExternalAddressSigner) Sign(addr iotago.Address, msg []byte) (signature iotago.Signature, err error) {
	return r.keyPair.Sign(addr, msg)
}
