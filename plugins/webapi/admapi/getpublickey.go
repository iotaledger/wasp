package admapi

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry" // TODO: [KP] It should be in the Plugins, not packages.
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

func addPublicKeyEndpoint(adm *echo.Group) {
	adm.GET("/"+client.GetPubKeyInfoRoute(":address"), handleGetPubKeyInfo)
}

func handleGetPubKeyInfo(c echo.Context) error {
	addr, err := address.FromBase58(c.Param("address"))
	chainID := coretypes.ChainID(addr)
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid address: %+v", c.Param("address")))
	}
	log.Debugw("GetDKShare", "addr", addr.String())
	ks, err := registry.DefaultRegistry().LoadDKShare(&chainID)
	log.Debugw("GetDKShare", "addr", addr.String(), "err", err, "ks", ks)
	if err != nil {
		return err
	}
	pubkeys := make([]string, len(ks.PublicShares))
	for i, pk := range ks.PublicShares {
		pkb, err := pk.MarshalBinary()
		if err != nil {
			return err
		}
		pubkeys[i] = base58.Encode(pkb)
	}
	pkm, err := ks.SharedPublic.MarshalBinary()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &client.PubKeyInfo{
		Address:      jsonable.NewAddress(&addr),
		N:            ks.N,
		T:            ks.T,
		Index:        *ks.Index,
		PubKeys:      pubkeys,
		PubKeyMaster: base58.Encode(pkm),
	})
}
