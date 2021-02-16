// +build ignore

package dwfclient

import (
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/webapi/model/statequery"
)

type DWFClient struct {
	*chainclient.Client
	contractHname coretypes.Hname
}

func NewClient(scClient *chainclient.Client, contractHname coretypes.Hname) *DWFClient {
	return &DWFClient{
		Client:        scClient,
		contractHname: contractHname,
	}
}

func (dwf *DWFClient) Donate(amount int64, feedback string) (*sctransaction.Transaction, error) {
	return dwf.PostRequest(
		dwf.contractHname,
		donatewithfeedback.RequestDonate,
		chainclient.PostRequestParams{
			Transfer: map[balance.Color]int64{balance.ColorIOTA: amount},
			ArgsRaw:  codec.MakeDict(map[string]interface{}{donatewithfeedback.VarReqFeedback: feedback}),
		},
	)
}

func (dwf *DWFClient) Withdraw(amount int64) (*sctransaction.Transaction, error) {
	return dwf.PostRequest(
		dwf.contractHname,
		donatewithfeedback.RequestWithdraw,
		chainclient.PostRequestParams{
			ArgsRaw: codec.MakeDict(map[string]interface{}{donatewithfeedback.VarReqWithdrawSum: amount}),
		},
	)
}

type Status struct {
	*chainclient.SCStatus

	NumRecords      uint32
	FirstDonated    time.Time
	LastDonated     time.Time
	MaxDonation     int64
	TotalDonations  int64
	LastRecordsDesc []*donatewithfeedback.DonationInfo
}

const maxRecordsToFetch = 15

func (dwf *DWFClient) FetchStatus() (*Status, error) {
	scStatus, results, err := dwf.FetchSCStatus(func(query *statequery.Request) {
		query.AddScalar(donatewithfeedback.VarStateMaxDonation)
		query.AddScalar(donatewithfeedback.VarStateTotalDonations)
		query.AddTLogSlice(donatewithfeedback.VarStateTheLog, 0, 0)
	})
	if err != nil {
		return nil, err
	}

	status := &Status{SCStatus: scStatus}

	status.MaxDonation, _ = results.Get(donatewithfeedback.VarStateMaxDonation).MustInt64()
	status.TotalDonations, _ = results.Get(donatewithfeedback.VarStateTotalDonations).MustInt64()
	logSlice := results.Get(donatewithfeedback.VarStateTheLog).MustTLogSliceResult()
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

	query := statequery.NewRequest()
	query.AddTLogSliceData(donatewithfeedback.VarStateTheLog, fromIdx, logSlice.LastIndex, true)
	res, err := dwf.StateQuery(query)
	if err != nil {
		return nil, err
	}
	status.LastRecordsDesc, err = decodeRecords(res.Get(donatewithfeedback.VarStateTheLog).MustTLogSliceDataResult())
	if err != nil {
		return nil, err
	}
	return status, nil
}

func decodeRecords(sliceData *statequery.TLogSliceDataResult) ([]*donatewithfeedback.DonationInfo, error) {
	ret := make([]*donatewithfeedback.DonationInfo, len(sliceData.Values))
	for i, data := range sliceData.Values {
		lr, err := collections.ParseRawLogRecord(data)
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
