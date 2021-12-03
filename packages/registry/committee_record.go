package registry

import (
	"fmt"

	"github.com/iotaledger/hive.go/marshalutil"
)

// CommitteeRecord represents committee information
// TODO optimize: no need to persists address in the structure
type CommitteeRecord struct {
	Address iotago.Address
	Nodes   []string // "host_addr:port"
}

// NewCommitteeRecord
func NewCommitteeRecord(addr iotago.Address, nodes ...string) *CommitteeRecord {
	ret := &CommitteeRecord{
		Address: addr,
		Nodes:   make([]string, len(nodes)),
	}
	copy(ret.Nodes, nodes)
	return ret
}

// CommitteeRecordFromMarshalUtil
func CommitteeRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*CommitteeRecord, error) {
	ret := &CommitteeRecord{}
	var err error
	ret.Address, err = iotago.AddressFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	numNodes, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	ret.Nodes = make([]string, numNodes)
	for i := uint16(0); i < numNodes; i++ {
		strSize, err := mu.ReadUint16()
		if err != nil {
			return nil, err
		}
		d, err := mu.ReadBytes(int(strSize))
		if err != nil {
			return nil, err
		}
		ret.Nodes[i] = string(d)
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
	return fmt.Sprintf("Committee(Address: %s Nodes:%+v)", rec.Address.Base58(), rec.Nodes)
}
