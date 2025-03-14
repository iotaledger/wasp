module github.com/iotaledger/wasp/tools/stardust-migration

go 1.23.0

toolchain go1.24.0

replace (
	github.com/ethereum/go-ethereum => github.com/iotaledger/go-ethereum v1.15.5-wasp1
	github.com/iotaledger/wasp => ../..
)

require (
	github.com/dgravesa/go-parallel v0.6.0
	github.com/ethereum/go-ethereum v1.15.5
	github.com/iotaledger/bcs-go v0.0.0-20250306194908-c8bf0ba9c5f4
	github.com/iotaledger/hive.go/kvstore v0.0.0-20240520064018-c635e5900894
	github.com/iotaledger/iota.go/v3 v3.0.0-rc.3
	github.com/iotaledger/wasp v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	github.com/nnikolash/wasp-types-exported v0.0.0-20250312155306-cba3d80c2d26
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/samber/lo v1.49.1
	github.com/slack-go/slack v0.16.0
	github.com/stretchr/testify v1.10.0
	github.com/urfave/cli/v2 v2.27.5
)

require github.com/google/uuid v1.6.0 // indirect

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.22.0 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.6 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/consensys/bavard v0.1.30 // indirect
	github.com/consensys/gnark-crypto v0.16.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240724233137-53bbb0ceb27a // indirect
	github.com/crate-crypto/go-kzg-4844 v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.4.3 // indirect
	github.com/ethereum/c-kzg-4844 v1.0.3 // indirect
	github.com/ethereum/go-verkle v0.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/iancoleman/orderedmap v0.3.0 // indirect
	github.com/iotaledger/grocksdb v1.7.5-0.20230220105546-5162e18885c7 // indirect
	github.com/iotaledger/hive.go/constraints v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/hive.go/crypto v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/hive.go/ds v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/hive.go/ierrors v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/hive.go/lo v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/hive.go/logger v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/runtime v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-rc.1.0.20240517131232-748f1ce3a2d2 // indirect
	github.com/iotaledger/hive.go/stringify v0.0.0-20240520064018-c635e5900894 // indirect
	github.com/iotaledger/iota.go v1.0.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.36.2 // indirect
	github.com/pasztorpisti/qs v0.0.0-20171216220353-8d6c33ee906c // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/petermattis/goid v0.0.0-20250303134427-723919f7f203 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/logging v0.2.3 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.21.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/supranational/blst v0.3.14 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/wollac/iota-crypto-demo v0.0.0-20221117162917-b10619eccb98 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.dedis.ch/kyber/v3 v3.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241223144023-3abc09e42ca8 // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
