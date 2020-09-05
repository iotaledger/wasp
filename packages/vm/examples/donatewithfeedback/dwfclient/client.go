package dwfclient

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
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

type WithdrawParams struct {
	Amount            int64
	WaitForCompletion bool
	PublisherHosts    []string
	PublisherQuorum   int
	Timeout           time.Duration
}

func (client *DWFClient) Withdraw(par WithdrawParams) (*sctransaction.Transaction, error) {
	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      client.nodeClient,
		SenderSigScheme: client.sigScheme,
		BlockParams: []apilib.RequestBlockParams{{
			TargetSCAddress: client.scAddress,
			RequestCode:     donatewithfeedback.RequestWithdraw,
			Vars: map[string]interface{}{
				donatewithfeedback.VarReqWithdrawSum: par.Amount,
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
	*waspapi.SCStatus

	NumRecords      uint32
	FirstDonated    time.Time
	LastDonated     time.Time
	MaxDonation     int64
	TotalDonations  int64
	LastRecordsDesc []*donatewithfeedback.DonationInfo
}

const maxRecordsToFetch = 15

func (client *DWFClient) FetchStatus() (*Status, error) {
	scStatus, results, err := waspapi.FetchSCStatus(client.nodeClient, client.waspApiHost, client.scAddress, func(query *stateapi.QueryRequest) {
		query.AddScalar(donatewithfeedback.VarStateMaxDonation)
		query.AddScalar(donatewithfeedback.VarStateTotalDonations)
		query.AddTLogSlice(donatewithfeedback.VarStateTheLog, 0, 0)
	})
	if err != nil {
		return nil, err
	}

	status := &Status{SCStatus: scStatus}

	status.MaxDonation, _ = results[donatewithfeedback.VarStateMaxDonation].MustInt64()
	status.TotalDonations, _ = results[donatewithfeedback.VarStateTotalDonations].MustInt64()
	logSlice := results[donatewithfeedback.VarStateTheLog].MustTLogSliceResult()
	if !logSlice.IsNotEmpty {
		// no records
		return status, nil
	}
	status.NumRecords = logSlice.LastIndex - logSlice.FirstIndex + 1
	status.FirstDonated = time.Unix(0, logSlice.Earliest)
	status.LastDonated = time.Unix(0, logSlice.Latest)

	fromIdx := uint32(0)
	if status.NumRecords > maxRecordsToFetch {
		fromIdx = logSlice.LastIndex - maxRecordsToFetch + 1
	}

	query := stateapi.NewQueryRequest(client.scAddress)
	query.AddTLogSliceData(donatewithfeedback.VarStateTheLog, fromIdx, logSlice.LastIndex, true)
	results, err = waspapi.QuerySCState(client.waspApiHost, query)
	if err != nil {
		return nil, err
	}
	status.LastRecordsDesc, err = decodeRecords(results[donatewithfeedback.VarStateTheLog].MustTLogSliceDataResult())
	if err != nil {
		return nil, err
	}
	return status, nil
}

func decodeRecords(sliceData *stateapi.TLogSliceDataResult) ([]*donatewithfeedback.DonationInfo, error) {
	ret := make([]*donatewithfeedback.DonationInfo, len(sliceData.Values))
	for i, data := range sliceData.Values {
		lr, err := kv.ParseRawLogRecord(data)
		if err != nil {
			return nil, err
		}
		ret[i], err = donatewithfeedback.DonationInfoFromBytes(lr.Data)
		if err != nil {
			return nil, err
		}
		ret[i].When = time.Unix(0, lr.Timestamp)
	}
	return ret, nil
}
