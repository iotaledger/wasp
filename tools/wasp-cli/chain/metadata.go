package chain

import (
	"context"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
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

type MetadataArgs struct {
	PublicURL     nilableString
	EvmJSONRPCURL nilableString
	EvmWSURL      nilableString

	ChainName        nilableString
	ChainDescription nilableString
	ChainWebsite     nilableString
}

/*
Sets the metadata for a given chain.

The idea is to enable the chain admin to:
 1. Persist an url to the Tangle which returns metadata about the chain, which can be consumed by 3rd party software (like Firefly).
 2. Configure alternative urls for the EVM JSON and Websocket RPC in case a load balancer is providing those connections on other locations.

Currently, there are three url parameters available which can be set: `PublicURL`, `EVMJsonRPCURL`, `EVMWSURL`.

The logic is as follows:

SetMetadata accepts the URLs mentioned above.

	If parameters are missing, they are ignored and will not change.
	If a parameter is empty, Wasp will fall back to the default values.
	If a parameter is not empty, the cli validates the url and sets it as is.
*/
func initMetadataCmd() *cobra.Command {
	var (
		node           string
		chainAliasName string
		withOffLedger  bool

		useCliURL    bool
		metadataArgs = MetadataArgs{}
	)

	cmd := &cobra.Command{
		Use:   "set-metadata",
		Short: "Updates the metadata urls for a given chain id",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chainAliasName = defaultChainFallback(chainAliasName)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			updateMetadata(ctx, client, node, chainAliasName, withOffLedger, useCliURL, metadataArgs)
			return nil
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainAliasName)

	cmd.Flags().BoolVarP(&withOffLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	cmd.Flags().BoolVarP(&useCliURL, "use-cli-url", "u", false, "use the configured cli wasp api url as public url (overrides --public-url)")
	cmd.Flags().Var(&metadataArgs.PublicURL, "public-url", "the url leading to chain metadata f.e. (https://chain.network/v1/chains/:chainID)")
	cmd.Flags().Var(&metadataArgs.EvmJSONRPCURL, "evm-rpc-url", "the public facing evm json rpc url")
	cmd.Flags().Var(&metadataArgs.EvmWSURL, "evm-ws-url", "the public facing evm websocket url")

	cmd.Flags().Var(&metadataArgs.ChainName, "name", "the chain name")
	cmd.Flags().Var(&metadataArgs.ChainDescription, "description", "the chain description")
	cmd.Flags().Var(&metadataArgs.ChainWebsite, "website", "the official project website of the chain")
	return cmd
}

func validateAndPush(target *string, value nilableString) {
	// If the value was not explicitly set, add nothing to the dictionary.
	if !value.IsSet() {
		return
	}

	*target = value.String()
}

func validateAndPushURL(target *string, urlValue nilableString) {
	// If the url was not explicitly set, add nothing to the dictionary.
	if !urlValue.IsSet() {
		return
	}

	// If the url is empty, force the default value
	if urlValue.String() == "" {
		*target = ""
		return
	}

	// If the url is longer than 0, treat it as an absolute url which gets validated before adding
	_, err := url.ParseRequestURI(urlValue.String())
	if err != nil {
		log.Fatal(err)
	}

	*target = urlValue.String()
}

func updateMetadata(ctx context.Context, client *apiclient.APIClient, node string, chainAliasName string, withOffLedger bool, useCliURL bool, metadataArgs MetadataArgs) {
	chainInfo, _, err := client.CorecontractsAPI.GovernanceGetChainInfo(ctx).Execute() //nolint:bodyclose // false positive
	if err != nil {
		log.Fatal("Chain not found")
	}

	publicURL := ""

	if useCliURL {
		apiURL := config.WaspAPIURL(node)
		chainPath, err2 := url.JoinPath(apiURL, "/v1/chain/")
		log.Check(err2)

		publicURL = chainPath
	} else {
		validateAndPushURL(&publicURL, metadataArgs.PublicURL)
	}

	// Use metadata from the chain info response, overwrite existing values with changes in the arguments
	validateAndPushURL(&chainInfo.Metadata.EvmJsonRpcURL, metadataArgs.EvmJSONRPCURL)
	validateAndPushURL(&chainInfo.Metadata.EvmWebSocketURL, metadataArgs.EvmWSURL)
	validateAndPush(&chainInfo.Metadata.Name, metadataArgs.ChainName)
	validateAndPush(&chainInfo.Metadata.Description, metadataArgs.ChainDescription)
	validateAndPushURL(&chainInfo.Metadata.Website, metadataArgs.ChainWebsite)

	// Map data to serialize to bytes
	chainMetadata := &isc.PublicChainMetadata{
		EVMJsonRPCURL:   chainInfo.Metadata.EvmJsonRpcURL,
		EVMWebSocketURL: chainInfo.Metadata.EvmWebSocketURL,
		Name:            chainInfo.Metadata.Name,
		Description:     chainInfo.Metadata.Description,
		Website:         chainInfo.Metadata.Website,
	}

	postRequest(ctx, client, chainAliasName, governance.FuncSetMetadata.Message(&publicURL, &chainMetadata), chainclient.PostRequestParams{
		GasBudget: iotaclient.DefaultGasBudget,
	}, withOffLedger)
}
