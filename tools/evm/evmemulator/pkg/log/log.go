// Package log provides logging functionality for the wasp-cli tool,
// supporting various log levels and output formats.
package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"text/tabwriter"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
	"github.com/spf13/cobra"
)

var (
	VerboseFlag bool
	DebugFlag   bool
	JSONFlag    bool

	hiveLogger log.Logger
)

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&VerboseFlag, "verbose", "", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "output debug information")
	rootCmd.PersistentFlags().BoolVarP(&JSONFlag, "json", "j", false, "json output")
}

func HiveLogger() log.Logger {
	if hiveLogger == nil {
		hiveLogger = log.NewLogger(log.WithLevel(slog.LevelInfo), log.WithOutput(os.Stdout))

		if DebugFlag {
			hiveLogger = log.NewLogger(log.WithLevel(slog.LevelDebug), log.WithOutput(os.Stdout))
		}
	}
	return hiveLogger
}

func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func Verbosef(format string, args ...interface{}) {
	if VerboseFlag {
		Printf(format, args...)
	}
}

func Info(args ...any) {
	s := fmt.Sprint(args...)
	fmt.Println(s)
}

func Fatal(args ...any) {
	s := fmt.Sprint(args...)
	Printf("error: %s\n", s)
	if DebugFlag {
		Printf("%s", debug.Stack())
	}
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	Printf("error: %s\n", s)
	if DebugFlag {
		Printf("%s", debug.Stack())
	}
	os.Exit(1)
}

func DefaultJSONFormatter(i interface{}) ([]byte, error) {
	return json.MarshalIndent(i, "", "  ")
}

type CLIOutput interface {
	AsText() (string, error)
}

type ExtendedCLIOutput interface {
	CLIOutput
	AsJSON() (string, error)
}

type ErrorModel struct {
	Error string
}

func (b *ErrorModel) AsText() (string, error) {
	return b.Error, nil
}

func GetCLIOutputText(outputModel CLIOutput) (string, error) {
	if JSONFlag {
		if output, ok := outputModel.(ExtendedCLIOutput); ok {
			return output.AsJSON()
		}

		jsonOutput, err := DefaultJSONFormatter(outputModel)
		return string(jsonOutput), err
	}

	return outputModel.AsText()
}

func ParseCLIOutputTemplate(output CLIOutput, templateDefinition string) (string, error) {
	tpl := template.Must(template.New("clioutput").Parse(templateDefinition))
	w := new(bytes.Buffer)
	err := tpl.Execute(w, output)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

func PrintCLIOutput(output CLIOutput) {
	outputText, err := GetCLIOutputText(output)
	Check(err)
	Printf("%s", outputText)
	// make sure we always end with newline
	if !strings.HasSuffix(outputText, "\n") {
		Printf("\n")
	}
}

func Check(err error, msg ...string) {
	if err == nil {
		return
	}

	errorModel := &ErrorModel{err.Error()}

	apiError, ok := apiextensions.AsAPIError(err)
	if ok {
		if strings.Contains(apiError.Error, "401") {
			errorModel = &ErrorModel{"unauthorized request: are you logged in? (wasp-cli login)"}
		} else {
			errorModel.Error = apiError.Error

			if apiError.DetailError != nil {
				errorModel.Error += "\n" + apiError.DetailError.Error + "\n" + apiError.DetailError.Message
			}
		}
	}

	if len(msg) > 0 {
		errorModel.Error = msg[0] + ": " + errorModel.Error
	}

	message, _ := GetCLIOutputText(errorModel)
	Fatalf("%v", message)
}

func PrintTable(header []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 0, 2, ' ', 0)

	fmt.Fprintf(w, strings.Join(makeSeparator(header), "\t")+"\n")
	fmt.Fprintf(w, strings.Join(header, "\t")+"\n")
	fmt.Fprintf(w, strings.Join(makeSeparator(header), "\t")+"\n")
	for _, row := range rows {
		fmt.Fprintf(w, strings.Join(row, "\t")+"\n")
	}
	w.Flush()
}

func makeSeparator(header []string) []string {
	ret := make([]string, len(header))
	for i, s := range header {
		ret[i] = strings.Repeat("-", len(s))
	}
	return ret
}

type TreeItem struct {
	K string
	V interface{}
}

func PrintTree(node interface{}, tab, tabwidth int) {
	indent := strings.Repeat(" ", tab)
	switch node := node.(type) {
	case []TreeItem:
		for _, item := range node {
			fmt.Printf("%s%s: ", indent, item.K)
			if s, ok := item.V.(string); ok {
				fmt.Printf("%s\n", s)
			} else {
				fmt.Print("\n")
				PrintTree(item.V, tab+tabwidth, tabwidth)
			}
		}
	case dict.Dict:
		if len(node) == 0 {
			fmt.Printf("%s(empty)", indent)
			return
		}
		tree := make([]TreeItem, 0, len(node))
		for k, v := range node {
			tree = append(tree, TreeItem{
				K: fmt.Sprintf("%q", string(k)),
				V: cryptolib.EncodeHex(v),
			})
		}
		PrintTree(tree, tab, tabwidth)
	case string:
		fmt.Printf("%s%s\n", indent, node)
	case isc.CallArguments:
		fmt.Printf("%s%s\n", indent, node)
		return
	default:
		panic(fmt.Sprintf("no handler of value of type %T", node))
	}
}
