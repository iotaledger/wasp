package chain

import (
	"context"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

/*
Used to distinguish between an empty flag value or an unset flag.

--rpc-url="" 		-> IsSet: true, Value: ""
--rpc-url="test" 	-> IsSet: true, Value: "test"
--					-> IsSet: false, Value: ""
*/
type nilableString struct {
	isSet bool
	value string
}

func (n *nilableString) Set(x string) error {
	n.value = x
	n.isSet = true
	return nil
}

func (n *nilableString) String() string {
	return n.value
}

func (n *nilableString) IsSet() bool {
	return n.isSet
}

func (n *nilableString) Type() string {
	return "string"
}

/*
Sets the metadata for a given chain.

The idea is to enable the chain owner to:
 1. Persist an url to the Tangle which returns metadata about the chain, which can be consumed by 3rd party software (like Firefly).
 2. Configure alternative urls for the EVM JSON and Websocket RPC in case a load balancer is providing those connections on other locations.

Currently, there are three url parameters available which can be set: `PublicURL`, `EVMJsonRPCUrl`, `EVMWSUrl`.

The logic is as follows:

SetMetadata accepts the URLs mentioned above.

	If parameters are missing, they are ignored and will not change.
	If a parameter is empty, Wasp will fall back to the default values.
	If a parameter is not empty, the cli validates the url and sets it as is.
*/
func initMetadataCmd() *cobra.Command {
	var (
		node          string
		chainName     string
		withOffLedger bool

		useCliURL     bool
		publicURL     nilableString
		evmJSONRpcURL nilableString
		evmWSUrl      nilableString
	)

	cmd := &cobra.Command{
		Use:   "set-metadata",
		Short: "Updates the metadata urls for a given chain id",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainName = defaultChainFallback(chainName)
			chainID := config.GetChain(chainName)

			updateMetadata(node, chainName, chainID, withOffLedger, useCliURL, publicURL, evmJSONRpcURL, evmWSUrl)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainName)

	cmd.Flags().BoolVarP(&withOffLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	cmd.Flags().BoolVarP(&useCliURL, "use-cli-url", "u", false, "use the configured cli wasp api url as public url (overrides --public-url)")
	cmd.Flags().Var(&publicURL, "public-url", "the chains public url")
	cmd.Flags().Var(&evmJSONRpcURL, "evm-rpc-url", "the public facing evm json rpc url")
	cmd.Flags().Var(&evmWSUrl, "evm-ws-url", "the public facing evm websocket url")

	return cmd
}

func validateAndPushURL(dict dict.Dict, key kv.Key, metadataURL nilableString) error {
	// If the url was not explicitly set, add nothing to the dictionary.
	if !metadataURL.IsSet() {
		return nil
	}

	// If the url is empty, force the default value
	if metadataURL.String() == "" {
		dict.Set(key, []byte{})
		return nil
	}

	// If the url is longer than 0, treat it as an absolute url which gets validated before adding
	_, err := url.ParseRequestURI(metadataURL.String())
	if err != nil {
		return err
	}

	dict.Set(key, []byte(metadataURL.String()))

	return nil
}

func updateMetadata(node string, chainName string, chainID isc.ChainID, withOffLedger bool, useCliURL bool, metadataURL nilableString, evmJSONRpcURL nilableString, evmWsURL nilableString) {
	client := cliclients.WaspClient(node)

	_, _, err := client.ChainsApi.GetChainInfo(context.Background(), chainID.String()).Execute() //nolint:bodyclose // false positive
	if err != nil {
		log.Fatal("Chain not found")
	}

	args := dict.Dict{}

	if useCliURL {
		apiURL := config.WaspAPIURL(node)
		chainPath, err := url.JoinPath(apiURL, "/v1/chains/", chainID.String())
		log.Check(err)

		args.Set(governance.ParamPublicURL, []byte(chainPath))
	} else {
		if err := validateAndPushURL(args, governance.ParamPublicURL, metadataURL); err != nil {
			log.Fatal(err)
		}
	}

	if err := validateAndPushURL(args, governance.ParamMetadataEVMJsonRPCURL, evmJSONRpcURL); err != nil {
		log.Fatal(err)
	}

	if err := validateAndPushURL(args, governance.ParamMetadataEVMWebSocketURL, evmWsURL); err != nil {
		log.Fatal(err)
	}

	params := chainclient.PostRequestParams{
		Args: args,
	}

	postRequest(node, chainName, governance.Contract.Name, governance.FuncSetMetadata.Name, params, withOffLedger, true)
}
