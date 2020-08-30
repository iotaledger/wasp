package dwfclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
	"time"
)

type DWFClient struct {
	nodeClient  nodeclient.NodeClient
	waspApiHost string
	scAddress   *address.Address
	sigScheme   signaturescheme.SignatureScheme
}

func NewClient(nodeClient nodeclient.NodeClient, waspApiHost string, scAddress *address.Address, sigScheme signaturescheme.SignatureScheme) *DWFClient {
	return &DWFClient{nodeClient, waspApiHost, scAddress, sigScheme}
}

type DonateParams struct {
	Amount            int64
	Feedback          string
	WaitForCompletion bool
	PublisherHosts    []string
	PublisherQuorum   int
	Timeout           time.Duration
}

func (client *DWFClient) Donate(par DonateParams) (*sctransaction.Transaction, error) {
	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      client.nodeClient,
		SenderSigScheme: client.sigScheme,
		BlockParams: []apilib.RequestBlockParams{{
			TargetSCAddress: client.scAddress,
			RequestCode:     donatewithfeedback.RequestDonate,
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: par.Amount,
			},
			Vars: map[string]interface{}{
				donatewithfeedback.VarReqFeedback: par.Feedback,
			}},
		},
		Post:                true,
		WaitForConfirmation: par.WaitForCompletion,
		WaitForCompletion:   par.WaitForCompletion,
		PublisherHosts:      par.PublisherHosts,
		PublisherQuorum:     par.PublisherQuorum,
		Timeout:             par.Timeout,
	})
}

type HarvestParams struct {
	Amount            int64
	WaitForCompletion bool
	PublisherHosts    []string
	PublisherQuorum   int
	Timeout           time.Duration
}

func (client *DWFClient) Harvest(par HarvestParams) (*sctransaction.Transaction, error) {
	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      client.nodeClient,
		SenderSigScheme: client.sigScheme,
		BlockParams: []apilib.RequestBlockParams{{
			TargetSCAddress: client.scAddress,
			RequestCode:     donatewithfeedback.RequestHarvest,
			Vars: map[string]interface{}{
				donatewithfeedback.VarReqHarvestSum: par.Amount,
			},
		}},
		Post:                true,
		WaitForConfirmation: par.WaitForCompletion,
		WaitForCompletion:   par.WaitForCompletion,
		PublisherHosts:      par.PublisherHosts,
		PublisherQuorum:     par.PublisherQuorum,
		Timeout:             par.Timeout,
	})
}

type Status struct {
	SCBalance map[balance.Color]int64
	FetchedAt time.Time

	NumRecords     uint32
	FirstDonated   time.Time
	LastDonated    time.Time
	MaxDonation    int64
	TotalDonations int64
	LastRecords    []*donatewithfeedback.DonationInfo
}

const maxRecordsToFetch = 50

func (client *DWFClient) FetchStatus() (*Status, error) {
	status := &Status{
		FetchedAt: time.Now().UTC(),
	}
	var err error
	if status.SCBalance, err = client.fetchSCBalance(); err != nil {
		return nil, err
	}
	if err = client.fetchLogInfo(status); err != nil {
		return nil, err
	}
	return status, nil
}

func (client *DWFClient) fetchSCBalance() (map[balance.Color]int64, error) {
	outs, err := client.nodeClient.GetAccountOutputs(client.scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}

func (client *DWFClient) fetchLogInfo(status *Status) error {
	query := stateapi.NewQueryRequest(client.scAddress)
	query.AddInt64(donatewithfeedback.VarStateMaxDonation)
	query.AddInt64(donatewithfeedback.VarStateTotalDonations)
	query.AddTLogSlice(donatewithfeedback.VarStateTheLog, 0, 0)

	results, err := waspapi.QuerySCState(client.waspApiHost, query)
	if err != nil {
		return err
	}
	status.MaxDonation, _, err = results[donatewithfeedback.VarStateMaxDonation].MustInt64()
	if err != nil {
		return err
	}
	status.TotalDonations, _, err = results[donatewithfeedback.VarStateTotalDonations].MustInt64()
	if err != nil {
		return err
	}
	logSlice := results[donatewithfeedback.VarStateTheLog].MustTLogSliceResult()
	if !logSlice.IsNotEmpty {
		// no records
		return nil
	}
	status.NumRecords = logSlice.LastIndex - logSlice.FirstIndex + 1
	status.FirstDonated = time.Unix(0, logSlice.Earliest)
	status.LastDonated = time.Unix(0, logSlice.Latest)
	fromIdx := uint32(0)
	if status.NumRecords > maxRecordsToFetch {
		fromIdx = logSlice.LastIndex - maxRecordsToFetch + 1
	}
	query = stateapi.NewQueryRequest(client.scAddress)
	query.AddTLogSliceData(donatewithfeedback.VarStateTheLog, fromIdx, logSlice.LastIndex)
	results, err = waspapi.QuerySCState(client.waspApiHost, query)
	if err != nil {
		return err
	}
	status.LastRecords, err = decodeRecords(results[donatewithfeedback.VarStateTheLog].MustTLogSliceDataResult())
	return nil
}

func decodeRecords(sliceData *stateapi.TLogSliceDataResult) ([]*donatewithfeedback.DonationInfo, error) {
	ret := make([]*donatewithfeedback.DonationInfo, len(sliceData.Values))
	for i, data := range sliceData.Values {
		lr, err := kv.ParseRawLogRecord(data)
		if err != nil {
			return nil, err
		}
		ret[i], err = donatewithfeedback.DonationInfoFromBytes(lr.Data)
		ret[i].When = time.Unix(0, lr.Timestamp)
	}
	return ret, nil
}
