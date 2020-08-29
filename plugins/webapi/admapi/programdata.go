package admapi

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/progmeta"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type ProgramMetadataJsonable struct {
	ProgramHash string `json:"program_hash"`
	Location    string `json:"location"`
	VMType      string `json:"vm_type"`
	Description string `json:"description"`
}

//----------------------------------------------------------
func HandlerPutProgramMetaData(c echo.Context) error {
	var req ProgramMetadataJsonable
	var err error

	if err := c.Bind(&req); err != nil {
		return misc.OkJsonErr(c, err)
	}

	rec := registry.ProgramMetadata{}

	if rec.ProgramHash, err = hashing.HashValueFromBase58(req.ProgramHash); err != nil {
		return misc.OkJsonErr(c, err)
	}
	rec.Location = req.Location
	rec.VMType = req.VMType
	rec.Description = req.Description

	// TODO it is always overwritten!

	if err = registry.SaveProgramMetadata(&rec); err != nil {
		return misc.OkJsonErr(c, err)
	}

	log.Infof("Program metadata record has been saved. Program hash: %s, description: %s, location: %s",
		rec.ProgramHash.String(), rec.Description, rec.Location)
	return misc.OkJsonErr(c, nil)
}

type GetProgramMetadataRequest struct {
	ProgramHash string `json:"program_hash"`
}

type GetProgramMetadataResponse struct {
	ProgramMetadataJsonable
	ExistsMetadata bool   `json:"exists_metadata"`
	ExistsCode     bool   `json:"exists_code"`
	Error          string `json:"err"`
}

func HandlerGetProgramMetadata(c echo.Context) error {
	var req GetProgramMetadataRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &GetProgramMetadataResponse{
			Error: err.Error(),
		})
	}
	md, err := progmeta.GetProgramMetadata(req.ProgramHash)
	if err != nil {
		return misc.OkJson(c, &GetProgramMetadataResponse{Error: err.Error()})
	}
	if md == nil {
		return misc.OkJson(c, &GetProgramMetadataResponse{})
	}

	return misc.OkJson(c, &GetProgramMetadataResponse{
		ProgramMetadataJsonable: ProgramMetadataJsonable{
			ProgramHash: md.ProgramHash.String(),
			Location:    md.Location,
			VMType:      md.VMType,
			Description: md.Description,
		},
		ExistsMetadata: true,
		ExistsCode:     md.CodeAvailable,
	})
}
