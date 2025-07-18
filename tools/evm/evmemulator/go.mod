module github.com/iotaledger/wasp/v2/tools/evm/evmemulator

go 1.23.8

toolchain go1.24.4

replace (
	github.com/ethereum/go-ethereum => github.com/iotaledger/go-ethereum v1.15.5-wasp1
	github.com/iotaledger/wasp/v2 => ../../../
	github.com/iotaledger/wasp/v2/tools/wasp-cli => ../../wasp-cli/
	go.dedis.ch/kyber/v3 => github.com/kape1395/kyber/v3 v3.0.14-0.20230124095845-ec682ff08c93 // branch: dkg-2suites
)

require (
	github.com/ethereum/go-ethereum v1.15.5
	github.com/iotaledger/wasp/v2 v2.0.0-00010101000000-000000000000
	github.com/iotaledger/wasp/v2/tools/wasp-cli v0.0.0-20230923193348-da186f5602e0
	github.com/spf13/cobra v1.9.1
)

require (
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	fortio.org/safecast v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.22.0 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.6 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/consensys/bavard v0.1.30 // indirect
	github.com/consensys/gnark-crypto v0.16.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240724233137-53bbb0ceb27a // indirect
	github.com/crate-crypto/go-kzg-4844 v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v28.0.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ethereum/c-kzg-4844 v1.0.3 // indirect
	github.com/ethereum/go-verkle v0.2.2 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/iancoleman/orderedmap v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/iotaledger/bcs-go v0.0.0-20250716100925-71f848cac593 // indirect
	github.com/iotaledger/grocksdb v1.7.5-0.20230220105546-5162e18885c7 // indirect
	github.com/iotaledger/hive.go/constraints v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/crypto v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/db v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/ds v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/ierrors v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/lo v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/log v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/runtime v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/serializer/v2 v2.0.0-rc.1.0.20250409140545-e1a365dbea74 // indirect
	github.com/iotaledger/hive.go/stringify v0.0.0-20250409140545-e1a365dbea74 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250303091104-876f3ea5145d // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.36.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/petermattis/goid v0.0.0-20250303134427-723919f7f203 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/logging v0.2.3 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/supranational/blst v0.3.14 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/testcontainers/testcontainers-go v0.35.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/wollac/iota-crypto-demo v0.0.0-20221117162917-b10619eccb98 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.dedis.ch/kyber/v3 v3.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241223144023-3abc09e42ca8 // indirect
	google.golang.org/grpc v1.67.3 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
