package validation

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/pmezard/go-difflib/difflib"
)

func EnsureEqual(name, oldStr, newStr string) {
	if oldStr != newStr {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:       difflib.SplitLines(oldStr),
			B:       difflib.SplitLines(newStr),
			Context: 2,
		})

		cli.DebugLogf("%v diff:\n%v", strings.Title(name), diff)

		oldStateFilePath := os.TempDir() + "/stardust-migration-old-state.txt"
		newStateFilePath := os.TempDir() + "/stardust-migration-new-state.txt"
		cli.DebugLogf("Writing old and new states to files %v and %v\n", oldStateFilePath, newStateFilePath)

		os.WriteFile(oldStateFilePath, []byte(oldStr), 0644)
		os.WriteFile(newStateFilePath, []byte(newStr), 0644)

		panic(fmt.Errorf("%v are NOT equal", strings.Title(name)))
	}
}
