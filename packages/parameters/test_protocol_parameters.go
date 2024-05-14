package parameters

const (
	// TestTokenSupply is a test token supply constant.
	// Do not use this constant outside of unit tests, instead, query it via a node.
	TestTokenSupply = 2_779_530_283_277_761
)

// TestProtoParas is an instance of iotago.ProtocolParameters for testing purposes. It contains a zero vbyte rent cost.
// Only use this var in testing. Do not modify or use outside unit tests.
var TestProtoParas = &ProtocolParameters{
	Version:     2,
	NetworkName: "TestJungle",
	Bech32HRP:   "tgl",
	MinPoWScore: 0,
	RentStructure: RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	},
	TokenSupply: TestTokenSupply,
}
