package admapi

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

func addPublicKeyEndpoint(server *echo.Group) {
	server.GET("/"+client.GetPubKeyInfoRoute(":address"), handleGetPubKeyInfo)
}

func handleGetPubKeyInfo(c echo.Context) error {
	addr, err := address.FromBase58(c.Param("address"))
	if err != nil {
		return httperrors.BadRequest(fmt.Sprintf("Invalid address: %+v", c.Param("address")))
	}
	log.Debugw("GetDKShare", "addr", addr.String())
	ks, exist, err := registry.GetDKShare(&addr)
	log.Debugw("GetDKShare", "addr", addr.String(), "err", err, "exist", exist, "ks", ks)
	if err != nil {
		return err
	}
	if !exist {
		return httperrors.NotFound(fmt.Sprintf("Public key not found for address %s", addr.String()))
	}
	pubkeys := make([]string, len(ks.PubKeys))
	for i, pk := range ks.PubKeys {
		pkb, err := pk.MarshalBinary()
		if err != nil {
			return err
		}
		pubkeys[i] = base58.Encode(pkb)
	}
	pkm, err := ks.PubKeyMaster.MarshalBinary()
	if err != nil {
		return err
	}
	return misc.OkJson(c, &client.PubKeyInfo{
		Address:      jsonable.NewAddress(&addr),
		N:            ks.N,
		T:            ks.T,
		Index:        ks.Index,
		PubKeys:      pubkeys,
		PubKeyMaster: base58.Encode(pkm),
	})
}
