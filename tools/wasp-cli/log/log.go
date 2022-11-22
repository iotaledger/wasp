package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var (
	VerboseFlag bool
	DebugFlag   bool
	JSONFlag    bool

	hiveLogger *logger.Logger
)

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&VerboseFlag, "verbose", "", false, "verbose")
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolVarP(&JSONFlag, "json", "j", false, "json output")
}

func HiveLogger() *logger.Logger {
	if hiveLogger == nil {
		loggerCfg := logger.Config{
			Level:             "info",
			Encoding:          "console",
			OutputPaths:       []string{"stdout"},
			DisableEvents:     true,
			DisableCaller:     true,
			DisableStacktrace: true,
			StacktraceLevel:   "panic",
		}
		if DebugFlag {
			loggerCfg.Level = "debug"
			loggerCfg.DisableCaller = false
			loggerCfg.DisableStacktrace = false
		}
		var err error
		hiveLogger, err = logger.NewRootLogger(loggerCfg)
		Check(err)
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

func addNL(s string) string {
	if s[len(s)-1] != '\n' {
		return s + "\n"
	}
	return s
}

func Fatalf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	if DebugFlag {
		panic(s)
	}
	Printf("error: " + addNL(s))
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
	var result bytes.Buffer
	err := tpl.Execute(&result, output)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func PrintCLIOutput(output CLIOutput) {
	outputText, err := GetCLIOutputText(output)
	Check(err)
	Printf("%s\n", outputText)
}

func Check(err error) {
	if err != nil {
		errorModel := &ErrorModel{err.Error()}
		message, _ := GetCLIOutputText(errorModel)
		Fatalf("%v", message)
	}
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
				V: fmt.Sprintf("0x%s", iotago.EncodeHex(v)),
			})
		}
		PrintTree(tree, tab, tabwidth)
	case string:
		fmt.Printf("%s%s\n", indent, node)
	default:
		panic(fmt.Sprintf("no handler of value of type %T", node))
	}
}
