package validation

import (
	"fmt"
	"os"
	"strings"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/samber/lo"
)

var ConcurrentValidation bool // Ugly? Yes. Do we care? No.

// This function is used to run a function in a goroutine IF ConcurrentValidation is set to true.
// If not - the function runs immediately.
// This function is added to allow switching between parallel and sequential execution for easier debugging of validations.
func Go(f func()) <-chan struct{} {
	done := make(chan struct{})

	if ConcurrentValidation {
		go func() {
			defer close(done)
			f()
		}()

		return done
	}

	f()
	close(done)
	return done
}

// GoAllAndWait runs all functions in parallel and waits for all of them to finish (see comment above).
func GoAllAndWait(f ...func()) {
	promises := make([]<-chan struct{}, len(f))
	for i := range f {
		promises[i] = Go(f[i])
	}

	for _, p := range promises {
		<-p
	}
}

func EnsureEqual(comparisonName, leftStr, rightStr string) {
	if leftStr == rightStr {
		return
	}

	cli.DebugLogf("Strings are NOT equal - retrieving diff...")

	diff := lo.Must(difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       difflib.SplitLines(leftStr),
		B:       difflib.SplitLines(rightStr),
		Context: 0,
	}))

	cli.DebugLogf("\n\n*********************  DIFF  **********************\n\n")
	diffLinesCount := strings.Count(diff, "\n")
	cli.DebugLogf("Diff lines count: %v", diffLinesCount)
	if diffLinesCount > 100 {
		cli.DebugLogf("Diff is too long, showing only preview")
		diffLines := strings.Split(diff, "\n")
		diffFirstHalf := strings.Join(diffLines[:len(diffLines)/2], "\n")
		diffLastHalf := strings.Join(diffLines[len(diffLines)/2:], "\n")
		cli.DebugLogf("%v diff PREVIEW:\n%v\n%v", strings.Title(comparisonName),
			utils.MultilinePreview(diffFirstHalf), utils.MultilinePreview(diffLastHalf))
	} else {
		cli.DebugLogf("%v diff:\n%v", strings.Title(comparisonName), diff)
	}
	cli.DebugLogf("\n\n***************************************************\n\n")

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

func OldReadOnlyKVStore(r old_kv.KVStoreReader) old_kv.KVStore {
	type oldReadOnlyKVStore struct {
		old_kv.KVStoreReader
		old_kv.KVWriter
	}

	return &oldReadOnlyKVStore{
		KVStoreReader: r,
		KVWriter:      nil,
	}
}

func NewReadOnlyKVStore(r kv.KVStoreReader) kv.KVStore {
	type newReadOnlyKVStore struct {
		kv.KVStoreReader
		kv.KVWriter
	}

	return &newReadOnlyKVStore{
		KVStoreReader: r,
		KVWriter:      nil,
	}
}
