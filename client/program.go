package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
)

const (
	PutProgramRoute = "program"
)

func GetProgramMetadataRoute(progHash string) string {
	return "program/" + progHash
}

type ProgramMetadata struct {
	VMType      string `json:"vm_type"`
	Description string `json:"description"`
}

type PutProgramRequest struct {
	ProgramMetadata
	Code []byte `json:"code"`
}

type PutProgramResponse struct {
	ProgramHash *hashing.HashValue `json:"program_hash"`
}

// PutProgramMetadata calls node to write program code and ProgramMetadata record
func (c *WaspClient) PutProgram(vmType string, description string, code []byte) (*hashing.HashValue, error) {
	req := PutProgramRequest{
		ProgramMetadata: ProgramMetadata{
			VMType:      vmType,
			Description: description,
		},
		Code: code,
	}
	res := &PutProgramResponse{}
	if err := c.do(http.MethodPost, AdminRoutePrefix+"/"+PutProgramRoute, req, res); err != nil {
		return nil, err
	}
	return res.ProgramHash, nil
}

// GetProgramMetadata calls node to get ProgramMetadata by program hash
func (c *WaspClient) GetProgramMetadata(progHash *hashing.HashValue) (*registry.ProgramMetadata, error) {
	res := &ProgramMetadata{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+GetProgramMetadataRoute(progHash.String()), nil, res); err != nil {
		return nil, err
	}
	return &registry.ProgramMetadata{
		ProgramHash: *progHash,
		VMType:      res.VMType,
		Description: res.Description,
	}, nil
}
