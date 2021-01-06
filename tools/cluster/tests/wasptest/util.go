// +build ignore

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

func CreateOrigin1SC(clu *cluster.Cluster, sc *cluster.Chain) error {
	//fmt.Printf("------------------------------   Test 3: create origin of 1 SC\n")

	tx, err := sc.CreateOrigin(clu.Level1Client)
	if err != nil {
		return err
	}

	fmt.Printf("++++++++++ created origin tx:\n%s\n", tx.String())

	ownerAddr := sc.OriginatorAddress()
	sh := tx.MustState().StateHash()
	fmt.Printf("[cluster] new origin tx: id: %s, state hash: %v, addr: %s owner: %s\n",
		tx.ID().String(), sh.String(), sc.Address, ownerAddr.String())

	err = clu.Level1Client.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return err
	}

	// wait for init request to be processed
	time.Sleep(2 * time.Second)

	return nil
}

func mintNewColoredTokens(wasps *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, amount int64) (*balance.Color, error) {
	tx, err := vtxbuilder.NewColoredTokensTransaction(wasps.Level1Client, sigScheme, amount)
	if err != nil {
		return nil, err
	}
	err = wasps.Level1Client.PostAndWaitForConfirmation(tx)
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

func SendRequestsNTimes(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, n int, reqs []waspapi.RequestSectionParams) error {
	for i := 0; i < n; i++ {
		err := SendRequests(clu, sigScheme, reqs)
		if err != nil {
			return err
		}
	}
	return nil
}

func SendSimpleRequest(
	clu *cluster.Cluster,
	sigScheme signaturescheme.SignatureScheme,
	reqParams waspapi.RequestSectionParams) error {

	tx, err := waspapi.CreateRequestTransaction(waspapi.CreateRequestTransactionParams{
		Level1Client:         clu.Level1Client,
		SenderSigScheme:      sigScheme,
		RequestSectionParams: []waspapi.RequestSectionParams{reqParams},
		Post:                 true,
	})
	if err != nil {
		return err
	}
	fmt.Printf("[cluster] posting request tx %s\n", tx.ID().String())
	return nil
}

func SendSimpleRequestMulti(clu *cluster.Cluster, sigScheme signaturescheme.SignatureScheme, reqParams []waspapi.RequestSectionParams) error {
	tx, err := waspapi.CreateRequestTransaction(waspapi.CreateRequestTransactionParams{
		Level1Client:         clu.Level1Client,
		SenderSigScheme:      sigScheme,
		RequestSectionParams: reqParams,
		Post:                 true,
	})
	if err != nil {
		return err
	}
	fmt.Printf("[cluster] posting request tx %s\n", tx.ID().String())
	return nil
}
