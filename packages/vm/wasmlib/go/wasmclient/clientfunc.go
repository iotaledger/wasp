package wasmclient

import "github.com/iotaledger/hive.go/crypto/ed25519"

type ClientFunc struct {
	svc      *Service
	keyPair  *ed25519.KeyPair
	transfer map[string]uint64
}

// Post sends a request to the smart contract service
// You can wait for the request to complete by using the returned Request
// as parameter to Service.WaitRequest()
func (f *ClientFunc) Post(funcHname uint32, args *Arguments) Request {
	keyPair := f.keyPair
	if keyPair == nil {
		keyPair = f.svc.keyPair
	}
	return f.svc.PostRequest(funcHname, args, f.transfer, keyPair)
}

// Sign optionally overrides the default keypair from the service
func (f *ClientFunc) Sign(keyPair *ed25519.KeyPair) {
	f.keyPair = keyPair
}

// Transfer optionally indicates which tokens to transfer as part of the request
// The color string can be a base58-encoded 32-byte color, or "IOTA"
// The tokens are presumed to be available in the signing account on the chain
func (f *ClientFunc) Transfer(transfer map[string]uint64) {
	f.transfer = transfer
}
