package service

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ChainBackend interface {
	PostRequest(keyPair *ed25519.KeyPair, transfer map[ledgerstate.Color]uint64, scName string, funName string, optSize int, params ...interface{}) error
	CallView(scName string, funName string, params ...interface{}) (dict.Dict, error)
}
