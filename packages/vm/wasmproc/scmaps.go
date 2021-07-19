// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMaps struct {
	ScSandboxObject
}

func NewScMaps(host *wasmhost.KvStoreHost) *ScMaps {
	a := &ScMaps{}
	a.host = host
	return a
}

func (a *ScMaps) GetObjectID(keyID, typeID int32) int32 {
	return GetArrayObjectID(a, keyID, typeID, func() WaspObject {
		return NewScDict(a.host, dict.New())
	})
}
