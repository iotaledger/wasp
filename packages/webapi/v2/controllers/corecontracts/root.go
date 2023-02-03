package corecontracts

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
	"github.com/iotaledger/wasp/packages/webapi/v2/params"
)

func (c *Controller) findContract(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	paramHName, err := params.DecodeHNameFromHNameHexString(e, "contract")
	if err != nil {
		return err
	}

	r, err := c.vmService.CallViewByChainID(chainID, root.Contract.Hname(), root.ViewFindContract.Hname(), codec.MakeDict(map[string]interface{}{
		root.ParamHname: codec.EncodeHname(paramHName),
	}))
	if err != nil {
		return fmt.Errorf("call view failed: %w", err)
	}

	if r[root.ParamContractRecData] == nil {
		return errors.New("contract not found")
	}

	contractRecord, err := root.ContractRecordFromBytes(r[root.ParamContractRecData])
	if err != nil {
		return fmt.Errorf("cannot decode contract record: %w", err)
	}

	result := models.ContractInfoResponse{
		Name:        contractRecord.Name,
		Description: contractRecord.Description,
		HName:       contractRecord.Hname().String(),
		ProgramHash: contractRecord.ProgramHash,
	}

	return e.JSON(http.StatusOK, result)
}

func (c *Controller) getContractRecords(e echo.Context) error {
	chainID, err := params.DecodeChainID(e)
	if err != nil {
		return err
	}

	recs, err := c.vmService.CallViewByChainID(chainID, root.Contract.Hname(), root.ViewGetContractRecords.Hname(), nil)
	if err != nil {
		return fmt.Errorf("call view failed: %w", err)
	}

	contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
	if err != nil {
		return fmt.Errorf("cannot decode contract records: %w", err)
	}

	contractRecords := make([]models.ContractInfoResponse, len(contracts))

	for k, v := range contracts {
		contractRecords[k] = models.ContractInfoResponse{
			Name:        v.Name,
			Description: v.Description,
			HName:       v.Hname().String(),
			ProgramHash: v.ProgramHash,
		}
	}

	return e.JSON(http.StatusOK, contractRecords)
}
