package chain

import (
	"context"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
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

type MetadataArgs struct {
	PublicUrl     nilableString
	EvmJsonRPCUrl nilableString
	EvmWSUrl      nilableString

	ChainName        nilableString
	ChainDescription nilableString
	ChainWebsite     nilableString
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
		node           string
		chainAliasName string
		withOffLedger  bool

		useCliUrl    bool
		metadataArgs = MetadataArgs{}
	)

	cmd := &cobra.Command{
		Use:   "set-metadata",
		Short: "Updates the metadata urls for a given chain id",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainAliasName = defaultChainFallback(chainAliasName)
			chainID := config.GetChain(chainAliasName)

			updateMetadata(node, chainAliasName, chainID, withOffLedger, useCliUrl, metadataArgs)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainAliasName)

	cmd.Flags().BoolVarP(&withOffLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	cmd.Flags().BoolVarP(&useCliUrl, "use-cli-url", "u", false, "use the configured cli wasp api url as public url (overrides --public-url)")
	cmd.Flags().Var(&metadataArgs.PublicUrl, "public-url", "the url leading to chain metadata f.e. (https://chain.network/v1/chains/:chainID)")
	cmd.Flags().Var(&metadataArgs.EvmJsonRPCUrl, "evm-rpc-url", "the public facing evm json rpc url")
	cmd.Flags().Var(&metadataArgs.EvmWSUrl, "evm-ws-url", "the public facing evm websocket url")

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

func validateAndPushUrl(target *string, metadataUrl nilableString) {
	// If the url was not explicitly set, add nothing to the dictionary.
	if !metadataUrl.IsSet() {
		return
	}

	// If the url is empty, force the default value
	if len(metadataUrl.String()) == 0 {
		*target = ""
		return
	}

	// If the url is longer than 0, treat it as an absolute url which gets validated before adding
	_, err := url.ParseRequestURI(metadataURL.String())
	if err != nil {
		log.Fatal(err)
	}

	*target = metadataUrl.String()
}

func updateMetadata(node string, chainAliasName string, chainID isc.ChainID, withOffLedger bool, useCliUrl bool, metadataArgs MetadataArgs) {
	client := cliclients.WaspClient(node)

	chainInfo, _, err := client.ChainsApi.GetChainInfo(context.Background(), chainID.String()).Execute() //nolint:bodyclose // false positive
	if err != nil {
		log.Fatal("Chain not found")
	}

	publicUrl := ""

	if useCliURL {
		apiURL := config.WaspAPIURL(node)
		chainPath, err := url.JoinPath(apiURL, "/v1/chains/", chainID.String())
		log.Check(err)

		publicUrl = chainPath
	} else {
		validateAndPushUrl(&publicUrl, metadataArgs.PublicUrl)
	}

	// Use metadata from the chain info response, overwrite existing values with changes in the arguments
	validateAndPushUrl(&chainInfo.Metadata.EvmJsonRpcUrl, metadataArgs.EvmJsonRPCUrl)
	validateAndPushUrl(&chainInfo.Metadata.EvmWebSocketUrl, metadataArgs.EvmWSUrl)
	validateAndPush(&chainInfo.Metadata.Name, metadataArgs.ChainName)
	validateAndPush(&chainInfo.Metadata.Description, metadataArgs.ChainDescription)
	validateAndPushUrl(&chainInfo.Metadata.Website, metadataArgs.ChainWebsite)

	// Map data to serialize to bytes
	chainMetadata := isc.ChainMetadata{
		EVMJsonRPCURL:   chainInfo.Metadata.EvmJsonRpcUrl,
		EVMWebSocketURL: chainInfo.Metadata.EvmWebSocketUrl,
		Name:            chainInfo.Metadata.Name,
		Description:     chainInfo.Metadata.Description,
		Website:         chainInfo.Metadata.Website,
	}

	params := chainclient.PostRequestParams{
		Args: dict.Dict{
			governance.ParamPublicURL: []byte(publicUrl),
			governance.ParamMetadata:  chainMetadata.Bytes(),
		},
	}

	postRequest(node, chainAliasName, governance.Contract.Name, governance.FuncSetMetadata.Name, params, withOffLedger, true)
}
