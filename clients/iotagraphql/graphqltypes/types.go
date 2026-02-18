package graphqltypes

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

var IotaCoinType CoinType = CoinType(iotago.MustNewResourceType("0x2::iota::IOTA").String())

type TransactionBytes struct {
	TxBytes iotago.Base64Data
}
