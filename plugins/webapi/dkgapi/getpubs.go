package dkgapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

// The POST handler implements 'adm/getpubs' API
// Parameters(see GetPubKeyInfoRequest struct):
//     Address:   address of the DKShare
// API responds with public info of DKShare:

func HandlerGetKeyPubInfo(c echo.Context) error {
	var req GetPubKeyInfoRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &GetPubKeyInfoResponse{
			Err: err.Error(),
		})
	}
	return misc.OkJson(c, GetKeyPubInfoReq(&req))
}

type GetPubKeyInfoRequest struct {
	Address string `json:"address"` //base58
}

type GetPubKeyInfoResponse struct {
	Address      string   `json:"address"` //base58
	N            uint16   `json:"n"`
	T            uint16   `json:"t"`
	Index        uint16   `json:"index"`
	PubKeys      []string `json:"pub_keys"`       // base58
	PubKeyMaster string   `json:"pub_key_master"` // base58
	Err          string   `json:"err"`
}

func GetKeyPubInfoReq(req *GetPubKeyInfoRequest) *GetPubKeyInfoResponse {
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return &GetPubKeyInfoResponse{Err: err.Error()}
	}
	log.Debugw("GetDKShare", "addr", addr.String())
	ks, exist, err := registry.GetDKShare(&addr)
	log.Debugw("GetDKShare", "addr", addr.String(), "err", err, "exist", exist, "ks", ks)

	if err != nil {
		return &GetPubKeyInfoResponse{Err: err.Error()}
	}
	if !exist {
		return &GetPubKeyInfoResponse{Err: "dkshare not found"}
	}
	pubkeys := make([]string, len(ks.PubKeys))
	for i, pk := range ks.PubKeys {
		pkb, err := pk.MarshalBinary()
		if err != nil {
			return &GetPubKeyInfoResponse{Err: err.Error()}
		}
		pubkeys[i] = base58.Encode(pkb)
	}
	pkm, err := ks.PubKeyMaster.MarshalBinary()
	if err != nil {
		return &GetPubKeyInfoResponse{Err: err.Error()}
	}
	return &GetPubKeyInfoResponse{
		Address:      req.Address,
		N:            ks.N,
		T:            ks.T,
		Index:        ks.Index,
		PubKeys:      pubkeys,
		PubKeyMaster: base58.Encode(pkm),
	}
}
