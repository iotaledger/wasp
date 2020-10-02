package admapi

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type ProgramMetadata struct {
	VMType      string `json:"vm_type"`
	Description string `json:"description"`
}

type PutProgramRequest struct {
	ProgramMetadata
	Code []byte `json:"code"`
}

type PutProgramResponse struct {
	ProgramHash string `json:"program_hash"`
	Error       string `json:"err"`
}

//----------------------------------------------------------
func HandlerPutProgram(c echo.Context) error {
	var req PutProgramRequest
	var err error

	if err := c.Bind(&req); err != nil {
		return c.JSONPretty(http.StatusBadRequest, &PutProgramResponse{Error: err.Error()}, " ")
	}

	if req.VMType == "" {
		return c.JSONPretty(http.StatusBadRequest, &PutProgramResponse{Error: "vm_type is required"}, " ")
	}
	if req.Description == "" {
		return c.JSONPretty(http.StatusBadRequest, &PutProgramResponse{Error: "description is required"}, " ")
	}
	if req.Code == nil || len(req.Code) == 0 {
		return c.JSONPretty(http.StatusBadRequest, &PutProgramResponse{Error: "code is required (base64-encoded binary data)"}, " ")
	}

	progHash, err := registry.SaveProgramCode(req.Code)
	if err != nil {
		return c.JSONPretty(http.StatusInternalServerError, &PutProgramResponse{Error: err.Error()}, " ")
	}

	md := &registry.ProgramMetadata{
		ProgramHash: progHash,
		VMType:      req.VMType,
		Description: req.Description,
	}

	// TODO it is always overwritten!
	if err = md.Save(); err != nil {
		return c.JSONPretty(http.StatusInternalServerError, &PutProgramResponse{Error: err.Error()}, " ")
	}

	log.Infof("Program metadata record has been saved. Program hash: %s, description: %s",
		md.ProgramHash.String(), md.Description)
	return misc.OkJson(c, &PutProgramResponse{ProgramHash: progHash.String()})
}

type GetProgramMetadataResponse struct {
	ProgramMetadata
	Error string `json:"err"`
}

func HandlerGetProgramMetadata(c echo.Context) error {
	progHash, err := hashing.HashValueFromBase58(c.Param("hash"))
	if err != nil {
		return c.JSONPretty(http.StatusBadRequest, &GetProgramMetadataResponse{Error: err.Error()}, " ")
	}

	md, err := registry.GetProgramMetadata(&progHash)
	if err != nil {
		return c.JSONPretty(http.StatusBadRequest, &GetProgramMetadataResponse{Error: err.Error()}, " ")
	}
	if md == nil {
		return c.JSONPretty(http.StatusNotFound, &GetProgramMetadataResponse{Error: "Not found"}, " ")
	}

	return misc.OkJson(c, &GetProgramMetadataResponse{
		ProgramMetadata: ProgramMetadata{
			VMType:      md.VMType,
			Description: md.Description,
		},
	})
}
