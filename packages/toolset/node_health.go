package toolset

import (
	"context"
	"fmt"
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/wasp/clients/apiextensions"
)

func nodeHealth(args []string) error {
	fs := configuration.NewUnsortedFlagSet("", flag.ContinueOnError)
	nodeURLFlag := fs.String(FlagToolNodeURL, "http://localhost:9090", "URL of the wasp node (optional)")

	fs.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", ToolNodeHealth)
		fs.PrintDefaults()
		fmt.Printf("\nexample: %s --%s %s\n",
			ToolNodeHealth,
			FlagToolNodeURL,
			"http://192.168.1.221:9090",
		)
	}

	if err := parseFlagSet(fs, args); err != nil {
		return err
	}

	client, err := apiextensions.WaspAPIClientByHostName(*nodeURLFlag)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(getGracefulStopContext(), 5*time.Second)
	defer cancel()

	_, err = client.DefaultApi.GetHealth(ctx).Execute()
	if err != nil {
		return err
	}

	fmt.Printf("Node (%s) is healthy.\n", *nodeURLFlag)

	return nil
}
