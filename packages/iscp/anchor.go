package iscp

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

type Anchor interface {
	ChainID() ChainID
	StateIndex() uint32
	StateCommitment() hashing.HashValue // temporary
}

type anchorImpl struct {
	*iotago.AliasOutput
}

func AnchorFromAliasOutput(o *iotago.AliasOutput) (Anchor, error) {
	if _, err := hashing.HashValueFromBytes(o.StateMetadata); err != nil {
		return nil, err
	}
	return &anchorImpl{
		AliasOutput: o,
	}, nil
}

func (a anchorImpl) ChainID() ChainID {
	panic("implement me")
}

func (a anchorImpl) StateIndex() uint32 {
	return a.AliasOutput.StateIndex
}

func (a anchorImpl) StateCommitment() hashing.HashValue {
	ret, err := hashing.HashValueFromBytes(a.AliasOutput.StateMetadata)
	if err != nil {
		panic(err)
	}
	return ret
}
