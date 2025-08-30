package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	scannerStartBuf = 64 * 1024
	scannerMaxBuf   = 1024 * 1024
)

// Regex that strips ANSI escape codes such as: \x1b[2m, \x1b[92m, \x1b[0m
var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// Regexes to extract fields
var (
	// testNumRe extracts the numeric test identifier from lines like: "... test=123 ...".
	testNumRe = regexp.MustCompile(`\btest=(\d+)\b`)
	// containerRe extracts the container ID from lines like: "... container=abcdef ...".
	containerRe = regexp.MustCompile(`\bcontainer=([a-z0-9]+)\b`)
	// simResultRe matches result lines from the simulator log, e.g.:
	//   "[  10/2013] PASSED tests/..."
	//   "[gw3] [  10/2013] FAILED tests/..."
	simResultRe = regexp.MustCompile(`^(?:\[gw[^\]]+\]\s+)?\[\s*(\d+)\s*/\s*\d+\]\s+(PASSED|FAILED)\s+(.*)$`)
)

const (
	statusPassed = "PASSED"
	statusFailed = "FAILED"
)

// Match EnqueueTransactions error arrays anywhere in a line, case-insensitive,
// and tolerant to minor variations like "error:" vs "err:".
var enqueueErrRe = regexp.MustCompile(`(?i)EnqueueTransactions\b[^\[]*?err(?:or)?:\s*(\[[^\]]*\])`)

// Run executes the analysis.
// - Reads test log, optionally prints cleaned lines, and maps test->container.
// - Parses simulator log into test entries.
// - Resolves client logs dir, extracts error arrays for failed tests, and writes summary.
// Run executes the analysis workflow:
//  1. Parse test.log to optionally print cleaned lines and map testNumber -> containerID.
//  2. Parse simulator log to obtain per-test results (status + name).
//  3. Resolve the client logs directory.
//  4. Write a joined report including container IDs and (for failures) client error arrays.
func Run(testLogPath, simLogPath, clientLogsDirOverride, outPath string, printClean bool) error {
	// 1) Read and clean test.log lines, printing cleaned lines if requested.
	//    Collect a mapping testNumber -> containerID from lines containing both.
	testToContainerID, err := parseTestToContainerMap(testLogPath, printClean)
	if err != nil {
		return fmt.Errorf("reading test log: %w", err)
	}

	// 2) Parse simulator log for test number, status, and name.
	if simLogPath == "" {
		return errors.New("simulator log not found; specify with -simlog")
	}
	results, err := parseSimulatorResults(simLogPath)
	if err != nil {
		return fmt.Errorf("reading simulator log: %w", err)
	}

	// 3) Determine client logs directory.
	clientLogsDir := clientLogsDirOverride
	if clientLogsDir == "" {
		clientLogsDir = findClientLogsDir(testLogPath, simLogPath)
	}

	// 4) Write output joined with container IDs and error arrays from client logs for failures.
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("ensure output dir: %w", err)
	}
	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	// use a larger buffered writer to reduce syscalls when writing big summaries
	w := bufio.NewWriterSize(out, 64*1024)

	// cache for containerID -> error array to avoid re-reading files repeatedly
	containerErrCache := map[string]string{}
	// error occurrence counter across tests (each test contributes at most once per error)
	errCounts := map[string]int{}
	for _, r := range results {
		containerID := testToContainerID[r.test]
		if r.status == statusFailed {
			errArray := findClientErrorArray(clientLogsDir, containerID, containerErrCache)
			// Count errors once per test using minimal repeating unit
			unit := normalizeErrorUnit(errArray)
			// Do not count empty arrays in the summary to avoid noise,
			// but keep them in per-test output.
			if unit != "" && unit != "[]" {
				errCounts[unit]++
			}
			// Format: <test> <status> <container> <array> <test-name>
			fmt.Fprintf(w, "%s %s %s %s %s\n", r.test, r.status, containerID, errArray, r.name)
			continue
		}
		// PASSED: <test> <status> <container> <test-name>
		fmt.Fprintf(w, "%s %s %s %s\n", r.test, r.status, containerID, r.name)
	}
	if len(errCounts) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Errors summary:")
		type kv struct {
			k string
			v int
		}
		var arr []kv
		for k, v := range errCounts {
			arr = append(arr, kv{k, v})
		}
		sort.Slice(arr, func(i, j int) bool {
			if arr[i].v != arr[j].v {
				return arr[i].v > arr[j].v
			}
			return arr[i].k < arr[j].k
		})
		for _, p := range arr {
			fmt.Fprintf(w, "%d %s\n", p.v, p.k)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}
	return nil
}

// parseTestToContainerMap reads the test log, optionally prints cleaned lines to stdout,
// and returns a mapping from test number to container ID.
func parseTestToContainerMap(testLogPath string, printClean bool) (map[string]string, error) {
	testToContainerID := make(map[string]string)
	err := withFile(testLogPath, func(f *os.File) error {
		scanner := newScanner(f)
		for scanner.Scan() {
			raw := scanner.Text()
			clean := ansiRe.ReplaceAllString(raw, "")
			if printClean {
				fmt.Println(clean)
			}
			if tn := firstGroup(testNumRe, clean); tn != "" {
				if containerID := firstGroup(containerRe, clean); containerID != "" {
					testToContainerID[tn] = containerID
				}
			}
		}
		return scanner.Err()
	})
	if err != nil {
		return nil, err
	}
	return testToContainerID, nil
}

// parseSimulatorLog extracts (test,status,name) tuples from the simulator log.
func parseSimulatorResults(simLogPath string) ([]testResult, error) {
	results := make([]testResult, 0, 128)
	err := withFile(simLogPath, func(f *os.File) error {
		scanner := newScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if m := simResultRe.FindStringSubmatch(line); m != nil {
				results = append(results, testResult{test: m[1], status: m[2], name: m[3]})
			}
		}
		return scanner.Err()
	})
	return results, err
}

// resolveClientLogsDir attempts to find the wasp-client logs directory using
// common locations relative to the provided logs.
func findClientLogsDir(testLogPath, simLogPath string) string {
	var candidates []string
	if simLogPath != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(simLogPath), "wasp-client"))
	}
	workspaceDir := filepath.Dir(testLogPath)
	candidates = append(candidates,
		filepath.Join(workspaceDir, "logs", "wasp-client"),
		filepath.Join(workspaceDir, "wasp-client"),
	)
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	// fallback to original guess to avoid empty path
	return filepath.Join(workspaceDir, "wasp-client")
}

// testResult holds a single test outcome parsed from the simulator log.
type testResult struct {
	test   string
	status string // PASSED or FAILED
	name   string
}

// withFile opens a file, calls fn with the opened *os.File, and closes it.
func withFile(path string, fn func(*os.File) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

// newScanner creates a *bufio.Scanner with a larger buffer suited to long lines.
func newScanner(f *os.File) *bufio.Scanner {
	s := bufio.NewScanner(f)
	buf := make([]byte, scannerStartBuf)
	s.Buffer(buf, scannerMaxBuf)
	return s
}

// firstGroup returns the first captured group for a regex, if any.
func firstGroup(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// findClientErrorArray finds and extracts the first EnqueueTransactions error array
// for the given containerID by scanning the corresponding wasp-client log file.
func findClientErrorArray(clientLogsDir, containerID string, cache map[string]string) string {
	if containerID == "" {
		return "[]"
	}
	if v, ok := cache[containerID]; ok {
		return v
	}
	// find client log by prefix
	var matchPath string
	prefix := "client-" + containerID
	// try fast glob first
	if candidates, _ := filepath.Glob(filepath.Join(clientLogsDir, prefix+"*")); len(candidates) > 0 {
		matchPath = candidates[0]
	} else {
		// fallback: walk the dir tree and stop at the first match
		_ = filepath.WalkDir(clientLogsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasPrefix(d.Name(), prefix) {
				matchPath = path
				return fs.SkipAll // stop walking entirely
			}
			return nil
		})
	}
	if matchPath == "" {
		cache[containerID] = "[]"
		return "[]"
	}
	// scan file for first matching line
	var errArray string = "[]"
	_ = withFile(matchPath, func(f *os.File) error {
		s := newScanner(f)
		for s.Scan() {
			line := s.Text()
			// strip any ANSI codes that may be present in client logs
			clean := ansiRe.ReplaceAllString(line, "")
			if m := enqueueErrRe.FindStringSubmatch(clean); m != nil {
				errArray = m[1]
				break
			}
		}
		return s.Err()
	})
	cache[containerID] = errArray
	return errArray
}

// normalizeErrorUnit normalizes an error slice string like "[msg msg]" or "[msg]" and
// returns a representative unit such that repeated identical messages within
// a test are collapsed to a single unit string.
func normalizeErrorUnit(arr string) string {
	s := strings.TrimSpace(arr)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "[]"
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return "[]"
	}
	norm := strings.Join(words, " ")
	// try to find shortest repeating word-prefix
	for i := 1; i <= len(words); i++ {
		if len(words)%i != 0 {
			continue
		}
		prefix := strings.Join(words[:i], " ")
		rep := len(words) / i
		var b strings.Builder
		for k := 0; k < rep; k++ {
			if k > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(prefix)
		}
		if b.String() == norm {
			return prefix
		}
	}
	return norm
}
