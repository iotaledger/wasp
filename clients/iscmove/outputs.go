package iscmove

import (
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// Related to: https://github.com/iotaledger/kinesis/blob/develop/crates/sui-framework/packages/stardust/sources/unlock_condition/storage_deposit_return_unlock_condition.move
type StorageDepositReturnUnlockCondition struct {
	ReturnAddress sui.Address
	ReturnAmount  uint64
}

// Related to: https://github.com/iotaledger/kinesis/blob/develop/crates/sui-framework/packages/stardust/sources/unlock_condition/timelock_unlock_condition.move
type TimelockUnlockCondition struct {
	UnixTime uint32
}

// Related to: https://github.com/iotaledger/kinesis/blob/develop/crates/sui-framework/packages/stardust/sources/unlock_condition/expiration_unlock_condition.move
type ExpirationUnlockCondition struct {
	Owner         sui.Address
	ReturnAddress sui.Address
	UnixTime      uint32
}

// Related to: https://github.com/iotaledger/kinesis/blob/isc-suijsonrpc/crates/sui-framework/packages/stardust/sources/basic/basic_output.move
type BasicOutput struct {
	ID   sui.ObjectID
	IOTA suijsonrpc.Balance
	// Does the "Balance" model really fit here, or should it rather be SafeSuiBigInt[uint64]?
	NativeTokens           Bag
	StorageDepositReturnUC *StorageDepositReturnUnlockCondition
	TimeLockUC             *TimelockUnlockCondition
	ExpirationUC           *ExpirationUnlockCondition
	Metadata               *[]uint8
	Tag                    *[]uint8
	Sender                 *sui.Address
}
