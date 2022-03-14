// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type Request struct {
	err error
	id  *iscp.RequestID
}

func (r Request) Error() error {
	return r.err
}
