package consensus

import (
	"sort"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"go.dedis.ch/kyber/v3/sign/tbls"
	"golang.org/x/xerrors"
)

type BatchProposal struct {
	ValidatorIndex          uint16
	StateOutputID           ledgerstate.OutputID
	RequestIDs              []coretypes.RequestID
	Timestamp               time.Time
	ConsensusManaPledge     identity.ID
	AccessManaPledge        identity.ID
	FeeDestination          *coretypes.AgentID
	SigShareOfStateOutputID tbls.SigShare
}

type consensusBatchParams struct {
	medianTs        time.Time
	accessPledge    identity.ID
	consensusPledge identity.ID
	feeDestination  *coretypes.AgentID
	entropy         hashing.HashValue
}

func BatchProposalFromBytes(data []byte) (*BatchProposal, error) {
	return BatchProposalFromMarshalUtil(marshalutil.New(data))
}

const errFmt = "BatchProposalFromMarshalUtil: %w"

func BatchProposalFromMarshalUtil(mu *marshalutil.MarshalUtil) (*BatchProposal, error) {
	ret := &BatchProposal{}
	var err error
	ret.ValidatorIndex, err = mu.ReadUint16()
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	ret.StateOutputID, err = ledgerstate.OutputIDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	ret.AccessManaPledge, err = identity.IDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	ret.ConsensusManaPledge, err = identity.IDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	ret.FeeDestination, err = coretypes.AgentIDFromMarshalUtil(mu)
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	ret.Timestamp, err = mu.ReadTime()
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	size, err := mu.ReadUint16()
	if err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	var sigShareSize byte
	if sigShareSize, err = mu.ReadByte(); err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	if ret.SigShareOfStateOutputID, err = mu.ReadBytes(int(sigShareSize)); err != nil {
		return nil, xerrors.Errorf(errFmt, err)
	}
	ret.RequestIDs = make([]coretypes.RequestID, size)
	for i := range ret.RequestIDs {
		ret.RequestIDs[i], err = coretypes.RequestIDFromMarshalUtil(mu)
		if err != nil {
			return nil, xerrors.Errorf(errFmt, err)
		}
	}
	return ret, nil
}

func (b *BatchProposal) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint16(b.ValidatorIndex).
		Write(b.StateOutputID).
		Write(b.AccessManaPledge).
		Write(b.ConsensusManaPledge).
		Write(b.FeeDestination).
		WriteTime(b.Timestamp).
		WriteUint16(uint16(len(b.RequestIDs))).
		WriteByte(byte(len(b.SigShareOfStateOutputID))).
		WriteBytes(b.SigShareOfStateOutputID)
	for i := range b.RequestIDs {
		mu.Write(b.RequestIDs[i])
	}
	return mu.Bytes()
}

// calcBatchParameters from a given ACS deterministically calculates timestamp, access and consensus
// mana pledges and fee destination
// Timestamp is calculated by taking closest value from above to the median.
// TODO final version of pladeges and fee destination
func (c *Consensus) calcBatchParameters(props []*BatchProposal) (*consensusBatchParams, error) {
	var retTS time.Time

	ts := make([]time.Time, len(props))
	for i := range ts {
		ts[i] = props[i].Timestamp
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Before(ts[j])
	})
	retTS = ts[len(props)/2]

	indices := make([]uint16, len(props))
	for i := range indices {
		indices[i] = uint16(i)
	}
	// verify signatures calculate entropy
	sigSharesToAggregate := make([][]byte, len(props))
	for i, prop := range props {
		err := c.committee.DKShare().VerifySigShare(c.stateOutput.ID().Bytes(), prop.SigShareOfStateOutputID)
		if err != nil {
			return nil, xerrors.Errorf("INVALID SIGNATURE in ACS from peer #%d: %v", prop.ValidatorIndex, err)
		}
		sigSharesToAggregate[i] = prop.SigShareOfStateOutputID
	}
	// aggregate signatures for use as unpredictable entropy
	signatureWithPK, err := c.committee.DKShare().RecoverFullSignature(sigSharesToAggregate, c.stateOutput.ID().Bytes())
	if err != nil {
		return nil, xerrors.Errorf("recovering signature from ACS: %v", err)
	}

	// selects pseudo-random based on seed, the calculated timestamp
	selectedIndex := util.SelectDeterministicRandomUint16(indices, retTS.UnixNano())
	return &consensusBatchParams{
		medianTs:        retTS,
		accessPledge:    props[selectedIndex].AccessManaPledge,
		consensusPledge: props[selectedIndex].ConsensusManaPledge,
		feeDestination:  props[selectedIndex].FeeDestination,
		entropy:         hashing.HashData(signatureWithPK.Bytes()),
	}, nil
}

// calcIntersection a simple algorithm to calculate acceptable intersection. It simply takes all requests
// seen by 1/3+1 node. The assumptions is there can be at max 1/3 of bizantine nodes, so if something is reported
// by more that 1/3 of nodes it means it is correct
func calcIntersection(acs []*BatchProposal, n uint16) []coretypes.RequestID {
	minNumberMentioned := n/3 + 1
	numMentioned := make(map[coretypes.RequestID]uint16)

	maxLen := 0
	for _, prop := range acs {
		for _, reqid := range prop.RequestIDs {
			s := numMentioned[reqid]
			numMentioned[reqid] = s + 1
		}
		if len(prop.RequestIDs) > maxLen {
			maxLen = len(prop.RequestIDs)
		}
	}
	ret := make([]coretypes.RequestID, 0, maxLen)
	for reqid, num := range numMentioned {
		if num >= minNumberMentioned {
			ret = append(ret, reqid)
		}
	}
	return ret
}
