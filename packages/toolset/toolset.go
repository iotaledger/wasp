package toolset

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

const (
	FlagToolNodeURL = "nodeURL"
	ToolNodeHealth  = "node-health"
)

// ShouldHandleTools checks if tools were requested.
func ShouldHandleTools() bool {
	args := os.Args[1:]

	for _, arg := range args {
		if strings.ToLower(arg) == "tool" || strings.ToLower(arg) == "tools" {
			return true
		}
	}

	return false
}

// HandleTools handles available tools.
func HandleTools() {
	args := os.Args[1:]
	if len(args) == 1 {
		listTools()
		os.Exit(1)
	}

	tools := map[string]func([]string) error{
		ToolNodeHealth: nodeHealth,
	}

	tool, exists := tools[strings.ToLower(args[1])]
	if !exists {
		fmt.Print("tool not found.\n\n")
		listTools()
		os.Exit(1)
	}

	if err := tool(args[2:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// help text was requested
			os.Exit(0)
		}

		fmt.Printf("\nerror: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func listTools() {
	fmt.Printf("%-20s queries the health endpoint of a wasp node\n", fmt.Sprintf("%s:", ToolNodeHealth))
}

func parseFlagSet(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Check if all parameters were parsed
	if fs.NArg() != 0 {
		return errors.New("too much arguments")
	}

	return nil
}

func getGracefulStopContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		<-gracefulStop
		cancel()
	}()

	return ctx
}
