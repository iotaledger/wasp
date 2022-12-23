// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"bytes"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
)

var emptyTransactionID = iotago.TransactionID{}

type OutputInfo struct {
	OutputID           iotago.OutputID
	Output             iotago.Output
	TransactionIDSpent iotago.TransactionID
}

func (o *OutputInfo) Consumed() bool {
	return o.TransactionIDSpent != emptyTransactionID
}

func NewOutputInfo(outputID iotago.OutputID, output iotago.Output, transactionIDSpent iotago.TransactionID) *OutputInfo {
	return &OutputInfo{
		OutputID:           outputID,
		Output:             output,
		TransactionIDSpent: transactionIDSpent,
	}
}

func (o *OutputInfo) AliasOutputWithID() *AliasOutputWithID {
	return NewAliasOutputWithID(o.Output.(*iotago.AliasOutput), o.OutputID)
}

type AliasOutputWithID struct {
	outputID    iotago.OutputID
	aliasOutput *iotago.AliasOutput
}

func NewAliasOutputWithID(aliasOutput *iotago.AliasOutput, outputID iotago.OutputID) *AliasOutputWithID {
	return &AliasOutputWithID{
		outputID:    outputID,
		aliasOutput: aliasOutput,
	}
}

func NewAliasOutputWithIDFromBytes(data []byte) (*AliasOutputWithID, error) {
	return NewAliasOutputWithIDFromMarshalUtil(marshalutil.New(data))
}

func NewAliasOutputWithIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (*AliasOutputWithID, error) {
	id, err := OutputIDFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}

	outputLen, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}

	outputBytes, err := mu.ReadBytes(int(outputLen))
	if err != nil {
		return nil, err
	}

	aliasOutput := &iotago.AliasOutput{}
	if _, err := aliasOutput.Deserialize(outputBytes, serializer.DeSeriModeNoValidation, nil); err != nil {
		return nil, err
	}

	return &AliasOutputWithID{
		outputID:    id,
		aliasOutput: aliasOutput,
	}, nil
}

func (a *AliasOutputWithID) Bytes() []byte {
	mu := marshalutil.New()
	mu = OutputIDToMarshalUtil(a.outputID, mu)
	outBytes, err := a.aliasOutput.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}
	return mu.WriteUint16(uint16(len(outBytes))).WriteBytes(outBytes).Bytes()
}

func (a *AliasOutputWithID) GetAliasOutput() *iotago.AliasOutput {
	return a.aliasOutput
}

func (a *AliasOutputWithID) OutputID() iotago.OutputID {
	return a.outputID
}

func (a *AliasOutputWithID) TransactionID() iotago.TransactionID {
	return a.outputID.TransactionID()
}

func (a *AliasOutputWithID) GetStateIndex() uint32 {
	return a.aliasOutput.StateIndex
}

func (a *AliasOutputWithID) GetStateMetadata() []byte {
	return a.aliasOutput.StateMetadata
}

func (a *AliasOutputWithID) GetStateAddress() iotago.Address {
	return a.aliasOutput.StateController()
}

func (a *AliasOutputWithID) GetAliasID() iotago.AliasID {
	return util.AliasIDFromAliasOutput(a.aliasOutput, a.outputID)
}

func (a *AliasOutputWithID) Equals(other *AliasOutputWithID) bool {
	if a != nil && other == nil {
		return false
	}
	out1, err := a.aliasOutput.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}
	out2, err := other.aliasOutput.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}
	return a.outputID == other.outputID && bytes.Equal(out1, out2)
}

func (a *AliasOutputWithID) Hash() hashing.HashValue {
	return hashing.HashDataBlake2b(a.Bytes())
}

func (a *AliasOutputWithID) String() string {
	if a == nil {
		return "nil"
	}
	return a.outputID.ToHex()
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

func OutputSetToOutputIDs(outputSet iotago.OutputSet) iotago.OutputIDs {
	outputIDs := make(iotago.OutputIDs, len(outputSet))
	i := 0
	for id := range outputSet {
		outputIDs[i] = id
		i++
	}
	return outputIDs
}
