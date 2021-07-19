package log

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	VerboseFlag bool
	DebugFlag   bool
)

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&VerboseFlag, "verbose", "", false, "verbose")
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "d", false, "debug")
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

func Check(err error) {
	if err != nil {
		Fatalf(err.Error())
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
