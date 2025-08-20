// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

// A rotation can be initiated by seting a target committee for the chain/node.
type inputRotateTo struct {
	address *iotago.Address
}

func NewInputRotateTo(address *iotago.Address) gpa.Input {
	return &inputRotateTo{address: address}
}

func (inp *inputRotateTo) String() string {
	return fmt.Sprintf("{cons.inputRotateTo: %s}", inp.address.ToHex())
}
