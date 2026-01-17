package graphqltypes

type ProtocolConfig struct {
	MaxSupportedProtocolVersion *BigInt                        `json:"maxSupportedProtocolVersion,omitempty"`
	MinSupportedProtocolVersion *BigInt                        `json:"minSupportedProtocolVersion,omitempty"`
	ProtocolVersion             *BigInt                        `json:"protocolVersion,omitempty"`
	Attributes                  map[string]ProtocolConfigValue `json:"attributes,omitempty"`
	FeatureFlags                map[string]bool                `json:"featureFlags,omitempty"`
}

type ProtocolConfigValue struct {
	U16 *uint16  `json:"u16,omitempty"`
	U32 *uint32  `json:"u32,omitempty"`
	U64 *uint64  `json:"u64,omitempty"`
	F64 *float64 `json:"f64,omitempty"`
}
