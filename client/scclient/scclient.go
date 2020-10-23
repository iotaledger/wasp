package scclient

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/url"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type SCClient struct {
	NodeClient nodeclient.NodeClient
	WaspClient *client.WaspClient
	Address    *address.Address
	SigScheme  signaturescheme.SignatureScheme

	timeout time.Duration

	publisherHost string
}

func New(
	nodeClient nodeclient.NodeClient,
	waspClient *client.WaspClient,
	scAddress *address.Address,
	sigScheme signaturescheme.SignatureScheme,
	waitForCompletionTimeout ...time.Duration,
) *SCClient {
	var t time.Duration
	if len(waitForCompletionTimeout) > 0 {
		t = waitForCompletionTimeout[0]
	}
	return &SCClient{
		NodeClient: nodeClient,
		WaspClient: waspClient,
		Address:    scAddress,
		SigScheme:  sigScheme,
		timeout:    t,
	}
}

func (sc *SCClient) PostRequest(
	code coretypes.EntryPointCode,
	mint map[address.Address]int64,
	transfer map[balance.Color]int64,
	vars map[string]interface{},
) (*sctransaction.Transaction, error) {
	if sc.timeout > 0 && len(sc.publisherHost) == 0 {
		info, err := sc.WaspClient.Info()
		if err != nil {
			return nil, err
		}
		u, err := url.Parse(sc.WaspClient.BaseURL())
		if err != nil {
			return nil, err
		}
		sc.publisherHost = fmt.Sprintf("%s:%d", u.Hostname(), info.PublisherPort)
	}

	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      sc.NodeClient,
		SenderSigScheme: sc.SigScheme,
		Mint:            mint,
		BlockParams: []apilib.RequestBlockParams{{
			TargetSCAddress: sc.Address,
			RequestCode:     code,
			Transfer:        transfer,
			Vars:            vars,
		}},
		Post:                true,
		WaitForConfirmation: sc.timeout > 0,
		WaitForCompletion:   sc.timeout > 0,
		PublisherHosts:      []string{sc.publisherHost},
		PublisherQuorum:     1,
		Timeout:             sc.timeout,
	})
}
