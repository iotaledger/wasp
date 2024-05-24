package sui

import (
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// Related to: https://github.com/iotaledger/kinesis/blob/develop/crates/sui-framework/packages/stardust/sources/unlock_condition/storage_deposit_return_unlock_condition.move
type StorageDepositReturnUnlockCondition struct {
	ReturnAddress sui_types.SuiAddress `json:"return_address"`
	ReturnAmount  uint64               `json:"return_amount"`
}

// Related to: https://github.com/iotaledger/kinesis/blob/develop/crates/sui-framework/packages/stardust/sources/unlock_condition/timelock_unlock_condition.move
type TimelockUnlockCondition struct {
	UnixTime uint32 `json:"unix_time"`
}

// Related to: https://github.com/iotaledger/kinesis/blob/develop/crates/sui-framework/packages/stardust/sources/unlock_condition/expiration_unlock_condition.move
type ExpirationUnlockCondition struct {
	Owner         sui_types.SuiAddress `json:"owner"`
	ReturnAddress sui_types.SuiAddress `json:"return_address"`
	UnixTime      uint32               `json:"unix_time"`
}

// Related to: https://github.com/iotaledger/kinesis/blob/isc-models/crates/sui-framework/packages/stardust/sources/basic/basic_output.move
type BasicOutput struct {
	ID                     sui_types.ObjectID                   `json:"id"`
	IOTA                   models.Balance                       `json:"balance"` // Does the "Balance" model really fit here, or should it rather be SafeSuiBigInt[uint64]?
	NativeTokens           Bag                                  `json:"native_tokens"`
	StorageDepositReturnUC *StorageDepositReturnUnlockCondition `json:"storage_deposit_return_uc"`
	TimeLockUC             *TimelockUnlockCondition             `json:"timelock_uc"`
	ExpirationUC           *ExpirationUnlockCondition           `json:"expiration_uc"`
	Metadata               *[]uint8                             `json:"metadata"`
	Tag                    *[]uint8                             `json:"tag"`
	Sender                 *sui_types.SuiAddress                `json:"sender"`
}
