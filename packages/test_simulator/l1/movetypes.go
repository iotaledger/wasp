package l1

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

var IotaCoinTypeStr = string(graphqltypes.IotaCoinType)

func CoinTypeString(innerType string) string {
	return fmt.Sprintf("%s::coin::Coin<%s>", iotago.IotaPackageIDIotaFramework.String(), innerType)
}

func BalanceTypeString(innerType string) string {
	return fmt.Sprintf("%s::balance::Balance<%s>", iotago.IotaPackageIDIotaFramework.String(), innerType)
}

func UpgradeCapTypeString() string {
	return fmt.Sprintf("%s::package::UpgradeCap", iotago.IotaPackageIDIotaFramework.String())
}

func ObjectIDTypeString() string {
	return fmt.Sprintf("%s::object::ID", iotago.IotaPackageIDIotaFramework.String())
}

func AsciiStringTypeString() string {
	return fmt.Sprintf("%s::ascii::String", iotago.IotaPackageIDMoveStdlib.String())
}

func ISCTypeString(packageID iotago.PackageID, module, name string) string {
	return fmt.Sprintf("%s::%s::%s", packageID.String(), module, name)
}
