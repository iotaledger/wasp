package multiclient

import (
	"fmt"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
)

// CheckProgramMetadata checks if metadata exists in hosts and is consistent
// return program meta data from the first host if all consistent, otherwise nil
func (m *MultiClient) CheckProgramMetadata(progHash *hashing.HashValue) (*registry.ProgramMetadata, error) {
	mdata := make([]*registry.ProgramMetadata, len(m.nodes))
	err := m.Do(func(i int, w *client.WaspClient) error {
		var err error
		mdata[i], err = w.GetProgramMetadata(progHash)
		return err
	})
	if err != nil {
		return nil, err
	}
	errInconsistent := fmt.Errorf("non existent or inconsistent program metadata for program hash %s", progHash.String())
	for _, md := range mdata {
		if !consistentProgramMetadata(mdata[0], md) {
			return nil, errInconsistent
		}
	}
	return mdata[0], nil
}

// consistentProgramMetadata does not check if code exists
func consistentProgramMetadata(md1, md2 *registry.ProgramMetadata) bool {
	if md1 == nil || md2 == nil {
		return false
	}
	if md1.ProgramHash != md2.ProgramHash {
		return false
	}
	if md1.VMType != md2.VMType {
		return false
	}
	if md1.Description != md2.Description {
		return false
	}
	return true
}
