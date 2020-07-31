package wasptest

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/cluster"
)

const goshimmerDirectly = true

func CreateOrigin1SC(clu *cluster.Cluster, sc *cluster.SmartContractFinalConfig) error {
	//fmt.Printf("------------------------------   Test 3: create origin of 1 SC\n")

	var bindAddress string
	if goshimmerDirectly {
		bindAddress = clu.Config.GoshimmerApiHost()
	} else {
		bindAddress = clu.ApiHosts()[0]
	}

	//fmt.Printf("++++++++++ create origin bind address: %s\n", bindAddress)

	tx, err := sc.CreateOrigin(bindAddress)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++++ created origin tx:\n%s\n", tx.String())

	ownerAddr := sc.OwnerAddress()
	sh := tx.MustState().StateHash()
	fmt.Printf("[cluster] new origin tx: id: %s, state hash: %v, addr: %s owner: %s\n",
		tx.ID().String(), sh.String(), sc.Address, ownerAddr.String())

	// in real situation we have to wait for confirmation
	err = clu.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return err
	}
	//fmt.Printf("[cluster] posted node origin tx to Goshimmer: addr: %s, txid: %s\n", sc.Address, tx.ID().String())

	return nil
}

func mintNewColoredTokens(wasps *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, amount int64) (*balance.Color, error) {
	tx, err := waspapi.NewColoredTokensTransaction(wasps.Config.GoshimmerApiHost(), sigScheme, amount)
	if err != nil {
		return nil, err
	}
	err = wasps.PostAndWaitForConfirmation(tx)
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

func SendRequestsNTimes(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, n int, reqs []*waspapi.RequestBlockJson, wait time.Duration) error {
	for i := 0; i < n; i++ {
		// in real situation one must wait until the previous request is confirmed
		// (because of access to the same owner address)
		err := SendRequests(clu, sigScheme, reqs)
		if err != nil {
			return err
		}
		time.Sleep(wait)
	}
	return nil
}

func SendSimpleRequest(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, reqParams waspapi.CreateSimpleRequestParams) error {
	tx, err := waspapi.CreateSimpleRequest(clu.Config.GoshimmerApiHost(), sigScheme, reqParams)
	if err != nil {
		return err
	}

	//fmt.Printf("[cluster] created request tx: %s\n", tx.String())
	//fmt.Printf("[cluster] posting tx: %s\n", tx.Transaction.String())

	err = clu.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		fmt.Printf("[cluster] posting tx: %s err = %v\n", tx.Transaction.String(), err)
		return err
	}
	//fmt.Printf("[cluster] posted request txid %s\n", tx.ID().String())
	return nil
}

func SendRequests(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, reqs []*waspapi.RequestBlockJson) error {
	tx, err := waspapi.CreateRequestTransaction(clu.Config.GoshimmerApiHost(), sigScheme, reqs)
	if err != nil {
		return err
	}

	//fmt.Printf("[cluster] created request tx: %s\n", tx.String())
	//fmt.Printf("[cluster] posting tx: %s\n", tx.Transaction.String())

	err = clu.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		fmt.Printf("[cluster] posting tx: %s err = %v\n", tx.Transaction.String(), err)
		return err
	}
	//fmt.Printf("[cluster] posted request txid %s\n", tx.ID().String())
	return nil
}
