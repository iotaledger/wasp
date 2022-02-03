// Package vmtxbuilder implements AnchorTransactionBuilder, a transaction builder used by the VM to construct
// anchor transaction. The AnchorTransactionBuilder keeps consistent state along operations of adding inputs
// and adding outputs (`Consume` and `AddOutput`).
// Total iotas available to on-chain accounts are kept in the anchor output.
// The builder automatically maintains `internal` outputs which holds on-chain total balances of native tokens: one UTXO
// for one non-zero balance of iotago.NativeTokenID.
// Whenever native tokens are moved to/form the chain, those internal UTXO are updated by consuming input/producing output.
// The builder automatically ensures necessary minimal dust deposit on each of internal outputs. For this, builder takes
// iotas from the total iotas on the chain or puts them back, depending on the needs of internal outputs.
// When txbuilder is unable to maintain consistent state, it panics. The following panic code are possible:
// - ErrProtocolExceptionInputLimitExceeded when maximum number of inputs in the transaction is exceeded
// - ErrProtocolExceptionOutputLimitExceeded when maximum number of outputs in the transaction is exceeded
// - ErrProtocolExceptionNumberOfNativeTokensLimitExceeded when number of total different tokenIDs is exceeded
// - ErrProtocolExceptionNotEnoughFundsForInternalDustDeposit when total number of iotas available is not enough for dust deposit of the new internal output
// - ErrNotEnoughIotaBalance attempt to debit more iotas than possible
// - ErrNotEnoughNativeAssetBalance attempt to debit more native tokens than possible
// - ErrOverflow overflow of arithmetics
package vmtxbuilder
