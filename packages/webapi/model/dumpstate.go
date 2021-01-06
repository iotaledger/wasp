package model

import "github.com/iotaledger/wasp/packages/kv/dict"

type SCStateDump struct {
	Index     uint32    `json:"index"`
	Variables dict.Dict `json:"variables"`
}
