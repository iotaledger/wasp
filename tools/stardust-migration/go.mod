module github.com/iotaledger/wasp/tools/stardust-migration

go 1.22

replace (
	github.com/ethereum/go-ethereum => github.com/iotaledger/go-ethereum v1.14.5-wasp1
	github.com/iotaledger/wasp => ../..
)

require (
	github.com/ethereum/go-ethereum v1.14.12
	github.com/iotaledger/hive.go/kvstore v0.0.0-20240319170702-c7591bb5f9f2
	github.com/iotaledger/iota.go/v3 v3.0.0-rc.3
	github.com/iotaledger/wasp v0.0.0-00010101000000-000000000000
	github.com/nnikolash/wasp-types-exported v0.0.0-20250127130851-9a5d0496de6f
	github.com/samber/lo v1.46.0
)

require (
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.0 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/getsentry/sentry-go v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/status-im/keycard-go v0.2.0 // indirect
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/btcsuite/btcutil v1.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.12.1 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240223125850-b1e8a79f509c // indirect
	github.com/crate-crypto/go-kzg-4844 v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.4.3 // indirect
	github.com/ethereum/c-kzg-4844 v1.0.0 // indirect
	github.com/ethereum/go-verkle v0.1.1-0.20240829091221-dffa7562dbe9 // indirect
	github.com/fardream/go-bcs v0.7.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/iancoleman/orderedmap v0.3.0 // indirect
	github.com/iotaledger/grocksdb v1.7.5-0.20230220105546-5162e18885c7 // indirect
	github.com/iotaledger/hive.go/constraints v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/crypto v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/ds v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/ierrors v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/lo v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/logger v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/runtime v0.0.0-20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-rc.1.0.20240319170702-c7591bb5f9f2 // indirect
	github.com/iotaledger/hive.go/stringify v0.0.0-20231106113411-94ac829adbb2 // indirect
	github.com/iotaledger/iota.go v1.0.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pasztorpisti/qs v0.0.0-20171216220353-8d6c33ee906c // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/petermattis/goid v0.0.0-20230904192822-1876fd5063bc // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/supranational/blst v0.3.13 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/wollac/iota-crypto-demo v0.0.0-20221117162917-b10619eccb98 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.dedis.ch/kyber/v3 v3.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240827150818-7e3bb234dfed // indirect
	google.golang.org/grpc v1.66.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
