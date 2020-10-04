package trclient

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/iotaledger/wasp/packages/subscribe"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

type TokenRegistryClient struct {
	nodeClient nodeclient.NodeClient
	waspHost   string
	scAddress  *address.Address
	sigScheme  signaturescheme.SignatureScheme
}

func NewClient(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address, sigScheme signaturescheme.SignatureScheme) *TokenRegistryClient {
	return &TokenRegistryClient{nodeClient, waspHost, scAddress, sigScheme}
}

type MintAndRegisterParams struct {
	Supply            int64           // number of tokens to mint
	MintTarget        address.Address // where to mint new Supply
	Description       string
	UserDefinedData   []byte
	WaitForCompletion bool
	PublisherHosts    []string
	PublisherQuorum   int
	Timeout           time.Duration // must be enough for confirmation of the request transaction processing of it (>20s)
}

func (trc *TokenRegistryClient) OwnerAddress() address.Address {
	return trc.sigScheme.Address()
}

// MintAndRegister mints new Supply of colored tokens to some address and sends request
// to register it in the TokenRegistry smart contract
func (trc *TokenRegistryClient) MintAndRegister(par MintAndRegisterParams) (*sctransaction.Transaction, error) {
	ownerAddr := trc.sigScheme.Address()
	outs, err := trc.nodeClient.GetConfirmedAccountOutputs(&ownerAddr)
	if err != nil {
		return nil, err
	}
	txb, err := txbuilder.NewFromOutputBalances(outs)
	if err != nil {
		return nil, err
	}
	err = txb.MintColor(par.MintTarget, balance.ColorIOTA, par.Supply)
	if err != nil {
		return nil, err
	}
	args := kv.NewMap()
	codec := args.Codec()
	codec.SetString(tokenregistry.VarReqDescription, par.Description)
	if par.UserDefinedData != nil {
		codec.Set(tokenregistry.VarReqUserDefinedMetadata, par.UserDefinedData)
	}

	reqBlk := sctransaction.NewRequestBlock(*trc.scAddress, tokenregistry.RequestMintSupply)
	reqBlk.SetArgs(args)
	err = txb.AddRequestBlock(reqBlk)
	if err != nil {
		return nil, err
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(trc.sigScheme)

	var subs *subscribe.Subscription
	if !par.WaitForCompletion {
		err = trc.nodeClient.PostTransaction(tx.Transaction)
		if err != nil {
			return nil, err
		}
		return tx, nil
	}
	subs, err = subscribe.SubscribeMulti(par.PublisherHosts, "request_out", par.PublisherQuorum)
	if err != nil {
		return nil, err
	}
	defer subs.Close()
	err = trc.nodeClient.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	if !subs.WaitForPattern([]string{"request_out", trc.scAddress.String(), tx.ID().String(), "0"}, par.Timeout) {
		return nil, fmt.Errorf("didnt't get confirmation message in %v", par.Timeout)
	}
	return tx, nil
}

type Status struct {
	*waspapi.SCStatus

	Registry                     map[balance.Color]*tokenregistry.TokenMetadata
	RegistrySortedByMintTimeDesc []*TokenMetadataWithColor // may be nil
}

type TokenMetadataWithColor struct {
	tokenregistry.TokenMetadata
	Color balance.Color
}

func (trc *TokenRegistryClient) FetchStatus(sortByAgeDesc bool) (*Status, error) {
	scStatus, results, err := waspapi.FetchSCStatus(trc.nodeClient, trc.waspHost, trc.scAddress, func(query *stateapi.QueryRequest) {
		query.AddDictionary(tokenregistry.VarStateTheRegistry, 100)
	})
	if err != nil {
		return nil, err
	}

	status := &Status{SCStatus: scStatus}

	status.Registry, err = decodeRegistry(results[tokenregistry.VarStateTheRegistry].MustDictionaryResult())
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

func decodeRegistry(result *stateapi.DictResult) (map[balance.Color]*tokenregistry.TokenMetadata, error) {
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
	query := stateapi.NewQueryRequest(trc.scAddress)
	query.AddDictionaryElement(tokenregistry.VarStateTheRegistry, color.Bytes())

	res, err := waspapi.QuerySCState(trc.waspHost, query)
	if err != nil {
		return nil, err
	}

	value := res.Queries[tokenregistry.VarStateTheRegistry].MustDictionaryElementResult()
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
