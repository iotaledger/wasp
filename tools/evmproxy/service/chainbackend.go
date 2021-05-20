package service

import (
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ChainBackend interface {
	PostRequest(scName string, funName string, optSize int, params ...interface{}) error
	CallView(scName string, funName string, params ...interface{}) (dict.Dict, error)
}
