package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

var (
	genesis          []string
	unlockedAccount  string
	listenAddr       string
	corsAllowOrigins []string
)

func main() {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Run:   start,
		Use:   "evmemulator",
		Short: "evmemulator runs an instance of the evmchain contract with Solo as backend",
		Long: fmt.Sprintf(`evmemulator runs an instance of the evmchain contract with Solo as backend.

evmemulator does the following:

- Starts a Solo environment (a framework for running local ISCP chains in-memory)
- Deploys an ISCP chain
- Deploys the evmchain ISCP contract (which runs an Ethereum chain on top of the ISCP chain)
- Starts a JSON-RPC server with the evmchain contract as backend

You can connect any Ethereum tool (eg Metamask) to this JSON-RPC server and use it for testing Ethereum contracts running on ISCP.

The default genesis allocation is: %s:%d
                                   private key: %s

By default the server has no unlocked accounts. To send transactions, either:

- use eth_sendRawTransaction
- configure an unlocked account with --account, and use eth_sendTransaction
`,
			evmtest.FaucetAddress,
			evmtest.FaucetSupply,
			hex.EncodeToString(crypto.FromECDSA(evmtest.FaucetKey)),
		),
	}

	log.Init(cmd)

	cmd.PersistentFlags().StringSliceVarP(&genesis, "genesis", "g", nil, "genesis allocation (format: <address>:<wei>,...)")
	cmd.PersistentFlags().StringVarP(&unlockedAccount, "account", "a", "", "unlocked account (hex-encoded private key)")
	cmd.PersistentFlags().StringVarP(&listenAddr, "listen", "l", ":8545", "JSON-RPC listen address")
	cmd.PersistentFlags().StringSliceVarP(&corsAllowOrigins, "cors", "c", []string{"*"}, "CORS allow origins")

	err := cmd.Execute()
	log.Check(err)
}

func start(cmd *cobra.Command, args []string) {
	env := solo.New(solo.NewFakeTestingT("evmproxy"), true, false)

	chainOwner, _ := env.NewKeyPairWithFunds()
	chain := env.NewChain(chainOwner, "iscpchain")
	err := chain.DeployContract(chainOwner, "evmchain", evmchain.Interface.ProgramHash,
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(getGenesis()),
	)
	log.Check(err)

	signer, _ := env.NewKeyPairWithFunds()

	backend := jsonrpc.NewSoloBackend(env, chain, signer)
	evmChain := jsonrpc.NewEVMChain(backend)

	accountManager := jsonrpc.NewAccountManager(getUnlockedAccount())

	rpcsrv := jsonrpc.NewServer(evmChain, accountManager)
	defer rpcsrv.Stop()

	serveHTTP(rpcsrv)
}

func getGenesis() core.GenesisAlloc {
	if len(genesis) == 0 {
		return core.GenesisAlloc{
			evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
		}
	}
	ret := core.GenesisAlloc{}
	for _, s := range genesis {
		parts := strings.Split(s, ":")
		addr := common.HexToAddress(parts[0])
		amount := big.NewInt(0)
		amount.SetString(parts[1], 10)
		ret[addr] = core.GenesisAccount{Balance: amount}
	}
	return ret
}

func getUnlockedAccount() []*ecdsa.PrivateKey {
	if unlockedAccount == "" {
		return nil
	}
	account, err := crypto.HexToECDSA(unlockedAccount)
	log.Check(err)
	return []*ecdsa.PrivateKey{account}
}

func serveHTTP(rpcsrv *rpc.Server) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
		Output: e.Logger.Output(),
	}))
	if log.DebugFlag {
		e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
			fmt.Printf("REQUEST:  %s\n", string(reqBody))
			fmt.Printf("RESPONSE: %s\n", string(resBody))
		}))
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: corsAllowOrigins,
		AllowMethods: []string{http.MethodPost, http.MethodGet},
		AllowHeaders: []string{"*"},
	}))
	e.Any("/", echo.WrapHandler(rpcsrv))

	fmt.Printf("Starting JSON-RPC server on %s\n", listenAddr)
	if err := e.Start(listenAddr); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err.Error())
		}
	}
}
