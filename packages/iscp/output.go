// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscp

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util"
)

type AliasOutputWithID struct {
	output *iotago.AliasOutput
	id     *iotago.UTXOInput
}

func NewAliasOutputWithID(output *iotago.AliasOutput, id *iotago.UTXOInput) *AliasOutputWithID {
	return &AliasOutputWithID{
		output: output,
		id:     id,
	}
}

func (a *AliasOutputWithID) GetAliasOutput() *iotago.AliasOutput {
	return a.output
}

func (a *AliasOutputWithID) ID() *iotago.UTXOInput {
	return a.id
}

func (a *AliasOutputWithID) OutputID() iotago.OutputID {
	return a.id.ID()
}

func (a *AliasOutputWithID) GetStateIndex() uint32 {
	return a.output.StateIndex
}

func (a *AliasOutputWithID) GetStateMetadata() []byte {
	return a.output.StateMetadata
}

func (a *AliasOutputWithID) GetStateAddress() iotago.Address {
	return a.output.StateController()
}

func (a *AliasOutputWithID) GetAliasID() iotago.AliasID {
	return util.AliasIDFromAliasOutput(a.output, a.id.ID())
}

func AliasOutputsEqual(ao1, ao2 *iotago.AliasOutput) bool {
	if ao1 == nil {
		return ao2 == nil
	}
	if ao2 == nil {
		return false
	}
	if ao1.Amount != ao2.Amount {
		return false
	}
	if !ao1.NativeTokens.Equal(ao2.NativeTokens) {
		return false
	}
	if ao1.AliasID != ao2.AliasID {
		return false
	}
	if ao1.StateIndex != ao2.StateIndex {
		return false
	}
	if !bytes.Equal(ao1.StateMetadata, ao2.StateMetadata) {
		return false
	}
	if ao1.FoundryCounter != ao2.FoundryCounter {
		return false
	}
	if len(ao1.Conditions) != len(ao2.Conditions) {
		return false
	}
	for index := range ao1.Conditions {
		if !ao1.Conditions[index].Equal(ao2.Conditions[index]) {
			return false
		}
	}
	return ao1.Features.Equal(ao2.Features)
}

func UTXOInputIDFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (*iotago.UTXOInput, error) {
	idBytes, err := marshalUtil.ReadBytes(iotago.OutputIDLength)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output ID: %v", err)
	}
	var oid iotago.OutputID
	copy(oid[:], idBytes)
	return oid.UTXOInput(), nil
}
