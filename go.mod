module github.com/iotaledger/wasp

go 1.16

require (
	github.com/NebulousLabs/errors v0.0.0-20181203160057-9f787ce8f69e // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20181203155948-6fb6489aac4e // indirect
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/anthdm/hbbft v0.0.0-20190702061856-0826ffdcf567
	github.com/bytecodealliance/wasmtime-go v0.21.0
	github.com/ethereum/go-ethereum v1.10.3
	github.com/iotaledger/goshimmer v0.7.2-0.20210628074845-4f90850164d1
	github.com/iotaledger/hive.go v0.0.0-20210625103722-68b2cf52ef4e
	github.com/knadh/koanf v0.15.0
	github.com/labstack/echo/v4 v4.1.13
	github.com/mr-tron/base58 v1.2.0
	github.com/pangpanglabs/echoswagger/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	go.dedis.ch/kyber/v3 v3.0.13
	go.nanomsg.org/mangos/v3 v3.0.1
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/anthdm/hbbft => github.com/kape1395/hbbft v0.0.0-20210712164915-258e54d010fb

replace go.dedis.ch/kyber/v3 v3.0.13 => github.com/kape1395/kyber/v3 v3.0.14-0.20210622094514-fefb81148dc3
