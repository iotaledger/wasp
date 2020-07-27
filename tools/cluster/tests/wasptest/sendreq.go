package wasptest

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/tools/cluster"
	"time"
)

func MakeRequests(n int, constr func(int) *waspapi.RequestBlockJson) []*waspapi.RequestBlockJson {
	ret := make([]*waspapi.RequestBlockJson, n)
	for i := range ret {
		ret[i] = constr(i)
	}
	return ret
}

func SendRequestsNTimes(clu *cluster.Cluster, senderIndexUtxodb int, n int, reqs []*waspapi.RequestBlockJson, wait time.Duration) error {
	for i := 0; i < n; i++ {
		// in real situation one must wait until the previous request is confirmed
		// (because of access to the same owner address)
		err := SendRequests(clu, senderIndexUtxodb, reqs)
		if err != nil {
			return err
		}
		time.Sleep(wait)
	}
	return nil
}

func SendSimpleRequest(clu *cluster.Cluster, senderIndexUtxodb int, reqParams waspapi.CreateSimpleRequestParams) error {
	tx, err := createSimpleRequestTx(clu.Config.GoshimmerApiHost(), senderIndexUtxodb, &reqParams)
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

func createSimpleRequestTx(node string, senderIndexUtxodb int, reqParams *waspapi.CreateSimpleRequestParams) (*sctransaction.Transaction, error) {
	sigScheme := utxodb.GetSigScheme(utxodb.GetAddress(senderIndexUtxodb))
	return waspapi.CreateSimpleRequest(node, sigScheme, *reqParams)
}

func SendRequests(clu *cluster.Cluster, senderIndexUtxodb int, reqs []*waspapi.RequestBlockJson) error {
	tx, err := createRequestTx(clu.Config.GoshimmerApiHost(), senderIndexUtxodb, reqs)
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

func createRequestTx(node string, senderIndexUtxodb int, reqs []*waspapi.RequestBlockJson) (*sctransaction.Transaction, error) {
	sigScheme := utxodb.GetSigScheme(utxodb.GetAddress(senderIndexUtxodb))
	return waspapi.CreateRequestTransaction(node, sigScheme, reqs)
}
