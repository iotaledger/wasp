package vmcontext

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

// contains methods which builds anchor transaction

// initTxBuilder creates anchor transaction builder for the block
func (vmctx *VMContext) initTxBuilder() {

}

func (vmctx *VMContext) BuildTransactionEssence(stateHash hashing.HashValue, timestamp time.Time) (*iotago.TransactionEssence, error) {
	//if err := vmctx.txBuilder.AddAliasOutputAsRemainder(vmctx.chainID.AsAddress(), stateHash[:]); err != nil {
	//	return nil, xerrors.Errorf("mustFinalizeRequestCall: %v", err)
	//}
	//tx, _, err := vmctx.txBuilder.WithTimestamp(timestamp).BuildEssence()
	//if err != nil {
	//	return nil, err
	//}
	//return tx, nil
	return nil, nil
}
