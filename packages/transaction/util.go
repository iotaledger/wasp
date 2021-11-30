package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

// GetAliasOutput return output or nil if not found
func GetAliasOutput(tx *iotago.Transaction, aliasAddr iotago.Address) *iotago.AliasOutput {
	return GetAliasOutputFromEssence(tx.Essence(), aliasAddr)
}

func GetAliasOutputFromEssence(essence *iotago.TransactionEssence, aliasAddr iotago.Address) *iotago.AliasOutput {
	for _, o := range essence.Outputs() {
		if out, ok := o.(*iotago.AliasOutput); ok {
			out1 := out.UpdateMintingColor().(*iotago.AliasOutput)
			if out1.GetAliasAddress().Equals(aliasAddr) {
				return out1
			}
		}
	}
	return nil
}
