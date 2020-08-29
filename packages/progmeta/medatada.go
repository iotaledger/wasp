package progmeta

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/examples"
)

type ProgramMetadata struct {
	ProgramHash   hashing.HashValue
	Location      string
	VMType        string
	Description   string
	CodeAvailable bool
}

// GetProgramMetadata return nil, nil if metadata does not exist
func GetProgramMetadata(progHashStr string) (*ProgramMetadata, error) {
	ph, err := hashing.HashValueFromBase58(progHashStr)
	if err != nil {
		return nil, err
	}
	proc, ok := examples.GetProcessor(progHashStr)
	if ok {
		return &ProgramMetadata{
			ProgramHash:   ph,
			Location:      "builtin",
			VMType:        "builtin",
			Description:   proc.GetDescription(),
			CodeAvailable: true,
		}, nil
	}
	md, exists, err := registry.GetProgramMetadata(&ph)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	_, exists, err = registry.GetProgramCode(&ph)
	if err != nil {
		return nil, err
	}
	return &ProgramMetadata{
		ProgramHash:   ph,
		Location:      md.Location,
		VMType:        md.VMType,
		Description:   md.Description,
		CodeAvailable: exists,
	}, nil
}
