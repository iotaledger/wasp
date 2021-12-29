package wasmclient

import "github.com/iotaledger/hive.go/crypto/ed25519"

type ClientFunc struct {
	svc      *Service
	keyPair  *ed25519.KeyPair
	transfer map[string]uint64
}

func (f *ClientFunc) Post(funcHname uint32, args *Arguments) Request {
	keyPair := f.keyPair
	if keyPair == nil {
		keyPair = f.svc.keyPair
	}
	return f.svc.PostRequest(funcHname, args, f.transfer, keyPair)
}

func (f *ClientFunc) Sign(keyPair *ed25519.KeyPair) {
	f.keyPair = keyPair
}
