package apilib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/packages/util/multicall"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
)

type CreateSCParams struct {
	Node                  nodeclient.NodeClient
	CommitteeApiHosts     []string
	CommitteePeeringHosts []string
	AccessNodes           []string
	N                     uint16
	T                     uint16
	OwnerSigScheme        signaturescheme.SignatureScheme
	ProgramHash           hashing.HashValue
	Description           string
	Textout               io.Writer
	Prefix                string
}

type ActivateSCParams struct {
	Addresses         []*address.Address
	ApiHosts          []string
	WaitForCompletion bool
	PublisherHosts    []string
	Timeout           time.Duration
}

type DeactivateSCParams struct {
	Addresses         []*address.Address
	ApiHosts          []string
	WaitForCompletion bool
	PublisherHosts    []string
	Timeout           time.Duration
}

func DeactivateSC(host string, addr *address.Address) error {
	return postScRequest(host, addr, "deactivate")
}

func ActivateSC(host string, addr *address.Address) error {
	return postScRequest(host, addr, "activate")
}

func postScRequest(host string, addr *address.Address, request string) error {
	addrStr := addr.String()
	url := fmt.Sprintf("http://%s/adm/sc/%s/%s", host, addrStr, request)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	var respbody misc.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&respbody)
	if err != nil {
		return fmt.Errorf("response status %d: %v", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || respbody.Error != "" {
		return fmt.Errorf("response status %d: %s", resp.StatusCode, respbody.Error)
	}
	return nil
}

func ActivateSCMulti(par ActivateSCParams) error {
	funs := make([]func() error, 0)
	for _, addr := range par.Addresses {
		for _, host := range par.ApiHosts {
			h := host
			addr1 := addr
			funs = append(funs, func() error {
				return ActivateSC(h, addr1)
			})
		}
	}
	if !par.WaitForCompletion {
		_, errs := multicall.MultiCall(funs, 1*time.Second)
		return multicall.WrapErrors(errs)
	}
	subs, err := subscribe.SubscribeMulti(par.PublisherHosts, "state")
	if err != nil {
		return err
	}
	defer subs.Close()
	_, errs := multicall.MultiCall(funs, 1*time.Second)
	err = multicall.WrapErrors(errs)
	if err != nil {
		return err
	}
	// SC is initialized when it reaches state index #1
	patterns := make([][]string, len(par.Addresses))
	for i := range patterns {
		patterns[i] = []string{"state", par.Addresses[i].String(), "1"}
	}
	succ := subs.WaitForPatterns(patterns, par.Timeout)
	if !succ {
		return fmt.Errorf("didn't receive activation message in %v", par.Timeout)
	}
	return nil
}

func DeactivateSCMulti(par DeactivateSCParams) error {
	funs := make([]func() error, 0)
	for _, addr := range par.Addresses {
		for _, host := range par.ApiHosts {
			h := host
			addr1 := addr
			funs = append(funs, func() error {
				return DeactivateSC(h, addr1)
			})
		}
	}
	if !par.WaitForCompletion {
		_, errs := multicall.MultiCall(funs, 1*time.Second)
		return multicall.WrapErrors(errs)
	}
	subs, err := subscribe.SubscribeMulti(par.PublisherHosts, "dismissed_committee")
	if err != nil {
		return err
	}
	defer subs.Close()
	_, errs := multicall.MultiCall(funs, 1*time.Second)
	err = multicall.WrapErrors(errs)
	if err != nil {
		return err
	}
	patterns := make([][]string, len(par.Addresses))
	for i := range patterns {
		patterns[i] = []string{"dismissed_committee", par.Addresses[i].String(), "1"}
	}
	succ := subs.WaitForPatterns(patterns, par.Timeout)
	if !succ {
		return fmt.Errorf("didn't receive deactivation message in %v", par.Timeout)
	}
	return nil
}

// CreateSC performs all actions needed to deploy smart contract, except activation
// noinspection ALL
func CreateSC(par CreateSCParams) (*address.Address, *balance.Color, error) {
	textout := ioutil.Discard
	if par.Textout != nil {
		textout = par.Textout
	}
	ownerAddr := par.OwnerSigScheme.Address()

	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "creating and deploying smart contract. Owner address is %s. Parameters N = %d, T = %d\n",
		ownerAddr.String(), par.N, par.T)
	// check if SC is hardcoded. If not, require consistent metadata in all nodes
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "checking program hash %s.. \n", par.ProgramHash.String())

	fmt.Fprint(textout, par.Prefix)
	md, err := CheckProgramMetadata(par.CommitteeApiHosts, &par.ProgramHash)
	if err != nil {
		fmt.Fprintf(textout, "checking program metadata: FAILED: %v\n", err)
		return nil, nil, err
	}
	fmt.Fprintf(textout, "checking program metadata: OK. location: '%s', VMType: '%s', description: '%s'\n",
		md.Location, md.VMType, md.Description)

	// generate distributed key set on committee nodes
	scAddr, err := GenerateNewDistributedKeySet(par.CommitteeApiHosts, par.N, par.T)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "generating distributed key set.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "generating distributed key set.. OK. Generated address = %s\n", scAddr.String())
	}

	allOuts, err := par.Node.GetConfirmedAccountOutputs(&ownerAddr)

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "requesting owner address' UTXOs from node.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "requesting owner address' UTXOs from node.. OK\n")
	}

	// create origin transaction
	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              *scAddr,
		OwnerSignatureScheme: par.OwnerSigScheme,
		AllInputs:            allOuts,
		ProgramHash:          par.ProgramHash,
		Description:          par.Description,
		InputColor:           balance.ColorIOTA,
	})

	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "creating origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "creating origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	err = par.Node.PostAndWaitForConfirmation(originTx.Transaction)
	fmt.Fprint(textout, par.Prefix)
	if err != nil {
		fmt.Fprintf(textout, "posting origin transaction.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprintf(textout, "posting origin transaction.. OK. Origin txid = %s\n", originTx.ID().String())
	}

	succ, errs := PutSCDataMulti(par.CommitteeApiHosts, registry.BootupData{
		Address:        *scAddr,
		OwnerAddress:   ownerAddr,
		Color:          (balance.Color)(originTx.ID()),
		CommitteeNodes: par.CommitteePeeringHosts,
		AccessNodes:    par.AccessNodes,
	})

	fmt.Fprint(textout, par.Prefix)
	if !succ {
		err = multicall.WrapErrors(errs)
		fmt.Fprintf(textout, "sending smart contract metadata to Wasp nodes.. FAILED: %v\n", err)
		return nil, nil, err
	} else {
		fmt.Fprint(textout, "sending smart contract metadata to Wasp nodes.. OK.\n")
	}
	// TODO not finished with access nodes

	scColor := (balance.Color)(originTx.ID())
	fmt.Fprint(textout, par.Prefix)
	fmt.Fprintf(textout, "smart contract has been created succesfully. Address: %s, Color: %s, N = %d, T = %d\n",
		scAddr.String(), scColor.String(), par.N, par.T)
	return scAddr, &scColor, err
}
