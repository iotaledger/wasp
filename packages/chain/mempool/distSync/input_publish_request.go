// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputPublishRequest struct {
	request isc.Request
}

func NewInputPublishRequest(request isc.Request) gpa.Input {
	return &inputPublishRequest{request: request}
}
