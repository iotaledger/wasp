package consensus1imp

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"golang.org/x/xerrors"
)

type batchProposal struct {
	ValidatorIndex      uint16
	StateOutputID       ledgerstate.OutputID
	RequestIDs          []coretypes.RequestID
	Timestamp           time.Time
	ConsensusManaPledge identity.ID
	AccessManaPledge    identity.ID
	FeeDestination      *coretypes.AgentID
}

func BatchProposalFromBytes(data []byte) (*batchProposal, error) {
	return BatchProposalFromMarshalUtil(marshalutil.New(data))
}

func BatchProposalFromMarshalUtil(mu *marshalutil.MarshalUtil) (*batchProposal, error) {
	ret := &batchProposal{}
	var err error
	ret.ValidatorIndex, err = mu.ReadUint16()
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	ret.StateOutputID, err = ledgerstate.OutputIDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	ret.AccessManaPledge, err = identity.IDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	ret.ConsensusManaPledge, err = identity.IDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	ret.FeeDestination, err = coretypes.AgentIDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	ret.Timestamp, err = mu.ReadTime()
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	size, err := mu.ReadUint16()
	if err != nil {
		return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
	}
	ret.RequestIDs = make([]coretypes.RequestID, size)
	for i := range ret.RequestIDs {
		ret.RequestIDs[i], err = coretypes.RequestIDFromMarshalUtil(mu)
		if err != nil {
			return nil, xerrors.Errorf("BatchProposalFromMarshalUtil: %w", err)
		}
	}
	return ret, nil
}

func (b *batchProposal) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint16(b.ValidatorIndex).
		Write(b.StateOutputID).
		Write(b.AccessManaPledge).
		Write(b.ConsensusManaPledge).
		Write(b.FeeDestination).
		WriteTime(b.Timestamp).
		WriteUint16(uint16(len(b.RequestIDs)))
	for i := range b.RequestIDs {
		mu.Write(b.RequestIDs[i])
	}
	return mu.Bytes()
}
