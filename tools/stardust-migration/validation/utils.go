package validation

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/pmezard/go-difflib/difflib"
)

func EnsureEqual(comparisonName, leftStr, rightStr string) {
	if leftStr == rightStr {
		return
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       difflib.SplitLines(leftStr),
		B:       difflib.SplitLines(rightStr),
		Context: 2,
	})

	cli.DebugLogf("%v diff:\n%v", strings.Title(comparisonName), diff)

	r := strings.NewReplacer(
		" ", "-",
		"(", "",
		")", "",
		",", "",
		".", "",
		":", "",
		";", "",
		"!", "",
		"?", "",
		"\"", "",
		"'", "",
		"`", "",
		"\\", "",
		"/", "",
		"|", "",
		"<", "",
		">", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"=", "",
		"+", "",
		"-", "",
		"*", "",
		"/", "",
		"%", "",
		"^", "",
		"&", "",
		"$", "",
		"#", "",
		"@", "",
		"~", "",
	)
	nameForFilePath := r.Replace(comparisonName)
	leftStateFilePath := fmt.Sprintf("%v/stardust-migration-%v-1.txt", os.TempDir(), nameForFilePath)
	rightStateFilePath := fmt.Sprintf("%v/stardust-migration-%v-2.txt", os.TempDir(), nameForFilePath)
	cli.DebugLogf("Writing compared strings to files %v and %v\n", leftStateFilePath, rightStateFilePath)

	if err := os.WriteFile(leftStateFilePath, []byte(leftStr), 0644); err != nil {
		cli.Logf("ERROR writing left string to file: %v: %v", leftStateFilePath, err)
	}

	if err := os.WriteFile(rightStateFilePath, []byte(rightStr), 0644); err != nil {
		cli.Logf("ERROR writing right string to file: %v: %v", rightStateFilePath, err)
	}

	panic(fmt.Errorf("%v are NOT equal", strings.Title(comparisonName)))
}
