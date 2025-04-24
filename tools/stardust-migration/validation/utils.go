package validation

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/iotaledger/wasp/packages/hashing"
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
func Go(f func()) (waitDone func()) {
	done := make(chan struct{})

	var panicErr error

	if ConcurrentValidation {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicErr = fmt.Errorf("%v:\n%v", r, string(debug.Stack()))
				}
			}()

			defer close(done)
			f()
		}()

		return func() {
			<-done
			if panicErr != nil {
				panic(panicErr)
			}
		}
	}

	f()
	close(done)
	return func() {}
}

// GoAllAndWait runs all functions in parallel and waits for all of them to finish (see comment above).
func GoAllAndWait(f ...func()) {
	promises := make([]func(), len(f))
	for i := range f {
		promises[i] = Go(f[i])
	}

	for _, wait := range promises {
		wait()
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

	var diffStr string

	diffStr = "\n\n*********************  DIFF  **********************\n\n"
	diffLinesCount := strings.Count(diff, "\n")
	cli.DebugLogf("Diff lines count: %v", diffLinesCount)
	if diffLinesCount > 100 {
		cli.DebugLogf("Diff is too long, showing only preview")
		diffLines := strings.Split(diff, "\n")
		diffFirstHalf := strings.Join(diffLines[:len(diffLines)/2], "\n")
		diffLastHalf := strings.Join(diffLines[len(diffLines)/2:], "\n")
		diffStr += fmt.Sprintf("%v diff PREVIEW:\n%v\n%v", strings.Title(comparisonName),
			utils.MultilinePreview(diffFirstHalf), utils.MultilinePreview(diffLastHalf))
	} else {
		diffStr += fmt.Sprintf("%v diff:\n%v", strings.Title(comparisonName), diff)
	}
	diffStr += "\n\n***************************************************\n\n"

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

	panic(fmt.Errorf("%v are NOT equal:\n%v", strings.Title(comparisonName), diffStr))
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

var HashValues = true

func hashValue[Value ~string | ~[]byte](v Value) string {
	if !HashValues {
		return string(v)
	}

	return hashing.HashData([]byte(v)).Hex()
}
