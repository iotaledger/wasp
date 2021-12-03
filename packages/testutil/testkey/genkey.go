package testkey

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/wasp/packages/util"
)

func GenKeyAddr(seedOpt ...util.Seed) (*ed25519.PrivateKey, iotago.Address) {
	var key ed25519.PrivateKey
	if len(seedOpt) > 0 {
		key = ed25519.NewKeyFromSeed(seedOpt[0][:])
	} else {
		key = util.NewPrivateKey()
	}
	addr := iotago.Ed25519AddressFromPubKey(key.Public().(ed25519.PublicKey))
	return &key, &addr
}
