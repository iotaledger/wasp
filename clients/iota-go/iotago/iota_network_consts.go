package iotago

var (
	IotaPackageIDMoveStdlib    = MustPackageIDFromHex("0x1")
	IotaPackageIDIotaFramework = MustPackageIDFromHex("0x2")
	IotaPackageIDIotaSystem    = MustPackageIDFromHex("0x3")
	IotaPackageIDBridge        = MustPackageIDFromHex("0xb")
	IotaPackageIDDeepbook      = MustPackageIDFromHex("0xdee9")
)

var (
	IotaObjectIDSystemState        = MustObjectIDFromHex("0x5")
	IotaObjectIDClock              = MustObjectIDFromHex("0x6")
	IotaObjectIDAuthenticatorState = MustObjectIDFromHex("0x7")
	IotaObjectIDRandomnessState    = MustObjectIDFromHex("0x8")
	IotaObjectIDBridge             = MustObjectIDFromHex("0x9")
	IotaObjectIDDenyList           = MustObjectIDFromHex("0x403")
)

var IotaSystemModuleName Identifier = "iota_system"

var (
	IotaSystemStateObjectSharedVersion        = SequenceNumber(1)
	IotaClockObjectSharedVersion              = SequenceNumber(1)
	IotaAuthenticatorStateObjectSharedVersion = SequenceNumber(1)
)
