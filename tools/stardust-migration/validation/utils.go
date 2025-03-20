package validation

import (
	"fmt"
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
		panic(fmt.Errorf("%v are NOT equal", strings.Title(name)))
	}
}
