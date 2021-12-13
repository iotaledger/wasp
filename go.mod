module github.com/iotaledger/wasp

go 1.16

require (
	github.com/NebulousLabs/errors v0.0.0-20181203160057-9f787ce8f69e // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20181203155948-6fb6489aac4e // indirect
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/anthdm/hbbft v0.0.0-20190702061856-0826ffdcf567
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/bygui86/multi-profile/v2 v2.1.0
	github.com/bytecodealliance/wasmtime-go v0.21.0
	github.com/ethereum/go-ethereum v1.10.13
	github.com/iotaledger/hive.go v0.0.0-20211123102045-85f0c1036bf7
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-20211208125510-04baae2057d6 // indirect
	github.com/iotaledger/iota.go/v3 v3.0.0-20211208200733-914ff5853384 // indirect
	github.com/knadh/koanf v1.2.1
	github.com/labstack/echo/v4 v4.2.1
	github.com/libp2p/go-libp2p v0.14.4
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-quic-transport v0.12.0
	github.com/libp2p/go-libp2p-tls v0.2.0
	github.com/libp2p/go-tcp-transport v0.2.4
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/pangpanglabs/echoswagger/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/second-state/WasmEdge-go v0.9.0-rc3
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/wasmerio/wasmer-go v1.0.4
	go.dedis.ch/kyber/v3 v3.0.13
	go.nanomsg.org/mangos/v3 v3.0.1
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.7
)

replace (
	github.com/iotaledger/iota.go/v3 => C:\Users\evaldas\Documents\proj\Go\src\github.com\lunfardo314\iota.go
	github.com/anthdm/hbbft => github.com/kape1395/hbbft v0.0.0-20210824083459-b949585b7515
	github.com/ethereum/go-ethereum => github.com/dessaya/go-ethereum v1.10.10-0.20211102133541-45878bcd628c
	github.com/linxGnu/grocksdb => github.com/gohornet/grocksdb v1.6.38-0.20211012114404-55f425442260
	go.dedis.ch/kyber/v3 v3.0.13 => github.com/kape1395/kyber/v3 v3.0.14-0.20210622094514-fefb81148dc3
)
