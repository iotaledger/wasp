package jsonrpc

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ChainBackend interface {
	PostOnLedgerRequest(keyPair *ed25519.KeyPair, scName string, funName string, transfer map[ledgerstate.Color]uint64, args dict.Dict) error
	PostOffLedgerRequest(keyPair *ed25519.KeyPair, scName string, funName string, transfer map[ledgerstate.Color]uint64, args dict.Dict) error
	CallView(scName string, funName string, args dict.Dict) (dict.Dict, error)
}
