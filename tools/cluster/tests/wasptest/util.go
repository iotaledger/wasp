package wasptest

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/txutil/vtxbuilder"
	"github.com/iotaledger/wasp/tools/cluster"
)

func CreateOrigin1SC(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) error {
	//fmt.Printf("------------------------------   Test 3: create origin of 1 SC\n")

	tx, err := sc.CreateOrigin(clu.NodeClient)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++++ created origin tx:\n%s\n", tx.String())

	ownerAddr := sc.OwnerAddress()
	sh := tx.MustState().StateHash()
	fmt.Printf("[cluster] new origin tx: id: %s, state hash: %v, addr: %s owner: %s\n",
		tx.ID().String(), sh.String(), sc.Address, ownerAddr.String())

	err = clu.NodeClient.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return err
	}

	// wait for init request to be processed
	time.Sleep(2 * time.Second)

	return nil
}

func mintNewColoredTokens(wasps *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, amount int64) (*balance.Color, error) {
	tx, err := vtxbuilder.NewColoredTokensTransaction(wasps.NodeClient, sigScheme, amount)
	if err != nil {
		return nil, err
	}
	err = wasps.NodeClient.PostAndWaitForConfirmation(tx)
	if err != nil {
		return nil, err
	}
	ret := (balance.Color)(tx.ID())

	fmt.Printf("[cluster] minted %d new tokens of color %s\n", amount, ret.String())
	//fmt.Printf("Value ts: %s\n", tx.String())
	return &ret, nil
}

func MakeRequests(n int, constr func(int) *waspapi.RequestBlockJson) []*waspapi.RequestBlockJson {
	ret := make([]*waspapi.RequestBlockJson, n)
	for i := range ret {
		ret[i] = constr(i)
	}
	return ret
}

func SendRequestsNTimes(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, n int, reqs []*waspapi.RequestBlockJson) error {
	for i := 0; i < n; i++ {
		err := SendRequests(clu, sigScheme, reqs)
		if err != nil {
			return err
		}
	}
	return nil
}

func SendSimpleRequest(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, reqParams waspapi.CreateSimpleRequestParamsOld) error {
	tx, err := waspapi.CreateSimpleRequestOld(clu.NodeClient, sigScheme, reqParams)
	if err != nil {
		return err
	}
	fmt.Printf("[cluster] posting request tx %s\n", tx.ID().String())
	return clu.PostTransaction(tx)
}

func SendSimpleRequestMulti(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, reqParams []waspapi.CreateSimpleRequestParamsOld) error {
	tx, err := waspapi.CreateSimpleRequestMultiOld(clu.NodeClient, sigScheme, reqParams)
	if err != nil {
		return err
	}
	fmt.Printf("[cluster] posting tx with %d requests: %s\n", len(reqParams), tx.ID().String())
	return clu.PostTransaction(tx)
}

func SendRequests(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, reqs []*waspapi.RequestBlockJson) error {
	tx, err := waspapi.CreateRequestTransactionOld(clu.NodeClient, sigScheme, reqs)
	if err != nil {
		return err
	}
	return clu.PostTransaction(tx)
}
