package iscclient

import (
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

var (
	cvt          wasmhost.WasmConvertor
	HrpForClient = iotago.NetworkPrefix("")
)

func clientBech32Decode(bech32 string) wasmtypes.ScAddress {
	hrp, addr, err := iotago.ParseBech32(bech32)
	if err != nil {
		panic(err)
	}
	if hrp != HrpForClient {
		panic("invalid protocol prefix: " + string(hrp))
	}
	return cvt.ScAddress(addr)
}

func clientBech32Encode(scAddress wasmtypes.ScAddress) string {
	addr := cvt.IscAddress(&scAddress)
	return addr.Bech32(HrpForClient)
}

func clientHashKeccak(buf []byte) wasmtypes.ScHash {
	h := sha3.NewLegacyKeccak256()
	h.Write(buf)
	return wasmtypes.HashFromBytes(h.Sum(nil))
}

func clientHashName(name string) wasmtypes.ScHname {
	h, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	_, err = h.Write([]byte(name))
	if err != nil {
		panic(err)
	}
	hash := h.Sum(nil)
	for i := 0; i < len(hash); i += 4 {
		ret := wasmtypes.HnameFromBytes(hash[i : i+4])
		if ret != 0 {
			return ret
		}
	}
	// astronomically unlikely to end up here
	return 1
}

func SetSandboxWrappers(chainID string) error {
	if HrpForClient != "" {
		return nil
	}

	// local client implementations for some sandbox functions
	wasmtypes.Bech32Decode = clientBech32Decode
	wasmtypes.Bech32Encode = clientBech32Encode
	wasmtypes.HashKeccak = clientHashKeccak
	wasmtypes.HashName = clientHashName

	// set the network prefix for the current network
	hrp, _, err := iotago.ParseBech32(chainID)
	if err != nil {
		return err
	}
	if HrpForClient != hrp && HrpForClient != "" {
		panic("WasmClient can only connect to one Tangle network per app")
	}
	HrpForClient = hrp
	return nil
}
