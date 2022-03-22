module github.com/iotaledger/wasp

go 1.16

require (
	github.com/NebulousLabs/errors v0.0.0-20181203160057-9f787ce8f69e // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20181203155948-6fb6489aac4e // indirect
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/anthdm/hbbft v0.0.0-20190702061856-0826ffdcf567
	github.com/bygui86/multi-profile/v2 v2.1.0
	github.com/bytecodealliance/wasmtime-go v0.34.0
	github.com/ethereum/go-ethereum v1.10.16
	github.com/iotaledger/goshimmer v0.8.5
	github.com/iotaledger/hive.go v0.0.0-20211207105259-9e48241c18f7
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-20220209164443-53ca2b8201b4
	github.com/iotaledger/iota.go/v3 v3.0.0-20220321174547-5b62a60c7774
	github.com/knadh/koanf v1.3.3
	github.com/labstack/echo/v4 v4.2.1
	github.com/libp2p/go-libp2p v0.15.0
	github.com/libp2p/go-libp2p-core v0.9.0
	github.com/libp2p/go-libp2p-quic-transport v0.12.0
	github.com/libp2p/go-libp2p-tls v0.2.0
	github.com/libp2p/go-tcp-transport v0.2.8
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.4.1
	github.com/pangpanglabs/echoswagger/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/second-state/WasmEdge-go v0.9.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/wasmerio/wasmer-go v1.0.4
	go.dedis.ch/kyber/v3 v3.0.13
	go.nanomsg.org/mangos/v3 v3.0.1
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20220322021311-435b647f9ef2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.7
)

replace (
	github.com/anthdm/hbbft => github.com/kape1395/hbbft v0.0.0-20210824083459-b949585b7515
	github.com/ethereum/go-ethereum => github.com/dessaya/go-ethereum v1.10.10-0.20220305060401-18f9e3da0f84
	//github.com/iotaledger/iota.go/v3 => C:\Users\evaldas\Documents\proj\Go\src\github.com\lunfardo314\iota.go
	github.com/iotaledger/goshimmer => github.com/kape1395/goshimmer v0.7.5-0.20220126105741-2bc797667497
	github.com/linxGnu/grocksdb => github.com/gohornet/grocksdb v1.6.38-0.20211012114404-55f425442260
	go.dedis.ch/kyber/v3 v3.0.13 => github.com/kape1395/kyber/v3 v3.0.14-0.20210622094514-fefb81148dc3
)
