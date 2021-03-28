package sctransaction

import "github.com/iotaledger/goshimmer/packages/ledgerstate"

// FindAliasOutput return output or nil if not found
func FindAliasOutput(tx *ledgerstate.TransactionEssence, aliasAddr ledgerstate.Address) *ledgerstate.AliasOutput {
	for _, o := range tx.Outputs() {
		if out, ok := o.(*ledgerstate.AliasOutput); ok && out.GetAliasAddress().Equals(aliasAddr) {
			return out
		}
	}
	return nil
}
