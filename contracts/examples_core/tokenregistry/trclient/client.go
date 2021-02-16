// +build ignore

package trclient

import (
	"bytes"
	"sort"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/webapi/model/statequery"
)

type TokenRegistryClient struct {
	*chainclient.Client
	contractHname coretypes.Hname
}

func NewClient(scClient *chainclient.Client, contractHname coretypes.Hname) *TokenRegistryClient {
	return &TokenRegistryClient{scClient, contractHname}
}

type MintAndRegisterParams struct {
	Supply          int64           // number of tokens to mint
	MintTarget      address.Address // where to mint new Supply
	Description     string
	UserDefinedData []byte
}

func (trc *TokenRegistryClient) OwnerAddress() address.Address {
	return trc.SigScheme.Address()
}

// MintAndRegister mints new Supply of colored tokens to some address and sends request
// to register it in the TokenRegistry smart contract
func (trc *TokenRegistryClient) MintAndRegister(par MintAndRegisterParams) (*sctransaction.Transaction, error) {
	args := make(map[string]interface{})
	args[tokenregistry.VarReqDescription] = par.Description
	if par.UserDefinedData != nil {
		args[tokenregistry.VarReqUserDefinedMetadata] = par.UserDefinedData
	}
	return trc.PostRequest(
		trc.contractHname,
		tokenregistry.RequestMintSupply,
		chainclient.PostRequestParams{
			Mint:    map[address.Address]int64{par.MintTarget: par.Supply},
			ArgsRaw: codec.MakeDict(args),
		},
	)
}

type Status struct {
	*chainclient.SCStatus

	Registry                     map[balance.Color]*tokenregistry.TokenMetadata
	RegistrySortedByMintTimeDesc []*TokenMetadataWithColor // may be nil
}

type TokenMetadataWithColor struct {
	tokenregistry.TokenMetadata
	Color balance.Color
}

func (trc *TokenRegistryClient) FetchStatus(sortByAgeDesc bool) (*Status, error) {
	scStatus, results, err := trc.FetchSCStatus(func(query *statequery.Request) {
		query.AddMap(tokenregistry.VarStateTheRegistry, 100)
	})
	if err != nil {
		return nil, err
	}

	status := &Status{SCStatus: scStatus}

	status.Registry, err = decodeRegistry(results.Get(tokenregistry.VarStateTheRegistry).MustMapResult())
	if err != nil {
		return nil, err
	}

	if !sortByAgeDesc {
		return status, nil
	}
	tslice := make([]*TokenMetadataWithColor, 0, len(status.Registry))
	for col, ti := range status.Registry {
		tslice = append(tslice, &TokenMetadataWithColor{
			TokenMetadata: *ti,
			Color:         col,
		})
	}
	sort.Slice(tslice, func(i, j int) bool {
		return tslice[i].Created > tslice[j].Created
	})
	status.RegistrySortedByMintTimeDesc = tslice
	return status, nil
}

func decodeRegistry(result *statequery.MapResult) (map[balance.Color]*tokenregistry.TokenMetadata, error) {
	registry := make(map[balance.Color]*tokenregistry.TokenMetadata)
	for _, e := range result.Entries {
		color, _, err := balance.ColorFromBytes(e.Key)
		if err != nil {
			return nil, err
		}
		tm := &tokenregistry.TokenMetadata{}
		if err := tm.Read(bytes.NewReader(e.Value)); err != nil {
			return nil, err
		}
		registry[color] = tm
	}
	return registry, nil
}

func (trc *TokenRegistryClient) Query(color *balance.Color) (*tokenregistry.TokenMetadata, error) {
	query := statequery.NewRequest()
	query.AddMapElement(tokenregistry.VarStateTheRegistry, color.Bytes())

	res, err := trc.StateQuery(query)
	if err != nil {
		return nil, err
	}

	value := res.Get(tokenregistry.VarStateTheRegistry).MustMapElementResult()
	if value == nil {
		// not found
		return nil, nil
	}

	tm := &tokenregistry.TokenMetadata{}
	if err := tm.Read(bytes.NewReader(value)); err != nil {
		return nil, err
	}

	return tm, nil
}
