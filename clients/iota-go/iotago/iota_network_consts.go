package iotago

var (
	IotaPackageIdMoveStdlib    = MustPackageIDFromHex("0x1")
	IotaPackageIdIotaFramework = MustPackageIDFromHex("0x2")
	IotaPackageIdIotaSystem    = MustPackageIDFromHex("0x3")
	IotaPackageIdBridge        = MustPackageIDFromHex("0xb")
	IotaPackageIdDeepbook      = MustPackageIDFromHex("0xdee9")
)

var (
	IotaObjectIdSystemState        = MustObjectIDFromHex("0x5")
	IotaObjectIdClock              = MustObjectIDFromHex("0x6")
	IotaObjectIdAuthenticatorState = MustObjectIDFromHex("0x7")
	IotaObjectIdRandomnessState    = MustObjectIDFromHex("0x8")
	IotaObjectIdBridge             = MustObjectIDFromHex("0x9")
	IotaObjectIdDenyList           = MustObjectIDFromHex("0x403")
)

var (
	IotaSystemModuleName Identifier = "iota_system"
)

var (
	IotaSystemStateObjectSharedVersion        = SequenceNumber(1)
	IotaClockObjectSharedVersion              = SequenceNumber(1)
	IotaAuthenticatorStateObjectSharedVersion = SequenceNumber(1)
)
