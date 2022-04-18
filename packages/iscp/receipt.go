package iscp

import "github.com/iotaledger/wasp/packages/vm/gas"

// Receipt represents a blocklog.RequestReceipt with translated error string
type Receipt struct {
	Request         []byte             `json:"request"`
	Error           *UnresolvedVMError `json:"error"`
	GasBudget       uint64             `json:"gasBudget"`
	GasBurned       uint64             `json:"gasBurned"`
	GasFeeCharged   uint64             `json:"gasFeeCharged"`
	BlockIndex      uint32             `json:"blockIndex"`
	RequestIndex    uint16             `json:"requestIndex"`
	TranslatedError string             `json:"translatedError"`
	GasBurnLog      *gas.BurnLog       `json:"-"`
}
