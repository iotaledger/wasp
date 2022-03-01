module github.com/iotaledger/wasp

go 1.16

require (
	github.com/NebulousLabs/errors v0.0.0-20181203160057-9f787ce8f69e // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20181203155948-6fb6489aac4e // indirect
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/anthdm/hbbft v0.0.0-20190702061856-0826ffdcf567
	github.com/bygui86/multi-profile/v2 v2.1.0
	github.com/bytecodealliance/wasmtime-go v0.32.0
	github.com/ethereum/go-ethereum v1.10.10
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/iotaledger/goshimmer v0.7.5-0.20210811162925-25c827e8326a
	github.com/iotaledger/hive.go v0.0.0-20210625103722-68b2cf52ef4e
	github.com/knadh/koanf v0.15.0
	github.com/labstack/echo/v4 v4.6.0
	github.com/libp2p/go-libp2p v0.14.4
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-quic-transport v0.12.0
	github.com/libp2p/go-libp2p-tls v0.2.0
	github.com/libp2p/go-tcp-transport v0.2.4
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/pangpanglabs/echoswagger/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/second-state/WasmEdge-go v0.9.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/wasmerio/wasmer-go v1.0.4
	go.dedis.ch/kyber/v3 v3.0.13
	go.nanomsg.org/mangos/v3 v3.0.1
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.7
)

replace (
	github.com/anthdm/hbbft => github.com/kape1395/hbbft v0.0.0-20210824083459-b949585b7515
	github.com/ethereum/go-ethereum => github.com/dessaya/go-ethereum v1.10.10-0.20211102133541-45878bcd628c
	github.com/iotaledger/goshimmer => github.com/kape1395/goshimmer v0.7.5-0.20220126105741-2bc797667497
	github.com/linxGnu/grocksdb => github.com/gohornet/grocksdb v1.6.38-0.20211012114404-55f425442260
	go.dedis.ch/kyber/v3 v3.0.13 => github.com/kape1395/kyber/v3 v3.0.14-0.20210622094514-fefb81148dc3
)
