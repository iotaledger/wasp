package committee_record //nolint:revive //TODO remove `_` from packagename

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/marshalutil"
)

// CommitteeRecord represents committee information
// TODO optimize: no need to persists address in the structure
type CommitteeRecord struct {
	Address ledgerstate.Address
	Nodes   []string // "host_addr:port"
}

// NewCommitteeRecord
func NewCommitteeRecord(addr ledgerstate.Address, nodes ...string) *CommitteeRecord {
	ret := &CommitteeRecord{
		Address: addr,
		Nodes:   make([]string, len(nodes)),
	}
	copy(ret.Nodes, nodes)
	return ret
}

// CommitteeRecordFromMarshalUtil
func CommitteeRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*CommitteeRecord, error) {
	ret := &CommitteeRecord{
		Nodes: make([]string, 0),
	}
	var err error
	ret.Address, err = ledgerstate.AddressFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	numNodes, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	for i := uint16(0); i < numNodes; i++ {
		strSize, err := mu.ReadUint16()
		if err != nil {
			return nil, err
		}
		d, err := mu.ReadBytes(int(strSize))
		if err != nil {
			return nil, err
		}
		ret.Nodes = append(ret.Nodes, string(d))
	}
	return ret, nil
}

// CommitteeRecordFromBytes
func CommitteeRecordFromBytes(data []byte) (*CommitteeRecord, error) {
	return CommitteeRecordFromMarshalUtil(marshalutil.New(data))
}

func (rec *CommitteeRecord) Bytes() []byte {
	mu := marshalutil.New().
		WriteBytes(rec.Address.Bytes()).
		WriteUint16(uint16(len(rec.Nodes)))
	for _, s := range rec.Nodes {
		b := []byte(s)
		mu.WriteUint16(uint16(len(b))).WriteBytes(b)
	}
	return mu.Bytes()
}

func (rec *CommitteeRecord) String() string {
	return fmt.Sprintf("Committee: %s. %+v", rec.Address.Base58(), rec.Nodes)
}
