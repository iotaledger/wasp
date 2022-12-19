package model

type OffLedgerRequestBody struct {
	Request Bytes `json:"request" swagger:"desc(Offledger Request (base64))"`
}
