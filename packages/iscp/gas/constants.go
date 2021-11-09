package gas

const (
	// Gas burned per 1 stored byte
	PerByte = 1
	// for the iscp.Sandbox.Utils interface
	UtilsHashingBlake2b              = 200
	UtilsHashingSha3                 = 300
	UtilsHashingHname                = 210
	UtilsBase58Encode                = 50
	UtilsBase58Decode                = 50
	UtilsED25519ValidSignature       = 500
	UtilsED25519AddressFromPublicKey = UtilsHashingBlake2b
	UtilsBLSValidSignature           = UtilsED25519ValidSignature * 40
	UtilsBLSAddressFromPublicKey     = UtilsHashingBlake2b
	UtilsBLSAggregateBLSSignature1   = 500
)
