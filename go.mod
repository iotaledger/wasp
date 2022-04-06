module github.com/iotaledger/wasp

go 1.16

require (
	github.com/NebulousLabs/errors v0.0.0-20181203160057-9f787ce8f69e // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20181203155948-6fb6489aac4e // indirect
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/VictoriaMetrics/fastcache v1.10.0 // indirect
	github.com/anthdm/hbbft v0.0.0-20190702061856-0826ffdcf567
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/bygui86/multi-profile/v2 v2.1.0
	github.com/bytecodealliance/wasmtime-go v0.35.0
	github.com/cockroachdb/errors v1.9.0 // indirect
	github.com/containerd/cgroups v1.0.3 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/elastic/gosigar v0.14.2 // indirect
	github.com/ethereum/go-ethereum v1.10.17
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/getsentry/sentry-go v0.13.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/huin/goupnp v1.0.3 // indirect
	github.com/iotaledger/goshimmer v0.8.11
	github.com/iotaledger/hive.go v0.0.0-20210625103722-68b2cf52ef4e
	github.com/ipfs/go-cid v0.1.0 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/klauspost/compress v1.15.1 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/klauspost/reedsolomon v1.9.16 // indirect
	github.com/knadh/koanf v1.4.0
	github.com/labstack/echo/v4 v4.7.2
	github.com/libp2p/go-libp2p v0.18.0
	github.com/libp2p/go-libp2p-core v0.14.0
	github.com/libp2p/go-libp2p-quic-transport v0.16.1
	github.com/libp2p/go-libp2p-tls v0.3.1
	github.com/libp2p/go-tcp-transport v0.5.1
	github.com/linxGnu/grocksdb v1.7.0 // indirect
	github.com/lucas-clemente/quic-go v0.27.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/miekg/dns v1.1.48 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-base32 v0.0.4 // indirect
	github.com/multiformats/go-multiaddr v0.5.0
	github.com/multiformats/go-multihash v0.1.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/pangpanglabs/echoswagger/v2 v2.4.0
	github.com/petermattis/goid v0.0.0-20220331194723-8ee3e6ded87a // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/common v0.33.0 // indirect
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/second-state/WasmEdge-go v0.9.2
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.1
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/wasmerio/wasmer-go v1.0.4
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.dedis.ch/kyber/v3 v3.0.13
	go.nanomsg.org/mangos/v3 v3.4.1
	go.uber.org/atomic v1.9.0
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29
	golang.org/x/net v0.0.0-20220403103023-749bd193bc2b // indirect
	golang.org/x/sys v0.0.0-20220405210540-1e041c57c461 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	lukechampine.com/blake3 v1.1.7 // indirect
	nhooyr.io/websocket v1.8.7
)

replace (
	github.com/anthdm/hbbft => github.com/kape1395/hbbft v0.0.0-20210824083459-b949585b7515
	github.com/ethereum/go-ethereum => github.com/dessaya/go-ethereum v1.10.10-0.20211102133541-45878bcd628c
	github.com/iotaledger/goshimmer => github.com/kape1395/goshimmer v0.7.5-0.20220126105741-2bc797667497
	github.com/linxGnu/grocksdb => github.com/gohornet/grocksdb v1.6.38-0.20211012114404-55f425442260
	go.dedis.ch/kyber/v3 v3.0.13 => github.com/kape1395/kyber/v3 v3.0.14-0.20210622094514-fefb81148dc3
)
