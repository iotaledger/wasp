package multiclient

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/registry_pkg/committee_record"
)

// PutChainRecord calls PutChainRecord in all wasp nodes
func (m *MultiClient) PutCommitteeRecord(bd *committee_record.CommitteeRecord) error {
	return m.Do(func(i int, w *client.WaspClient) error {
		return w.PutCommitteeRecord(bd)
	})
}
