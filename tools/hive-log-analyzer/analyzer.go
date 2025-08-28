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

// Regex to strip ANSI escape codes like \x1b[2m, \x1b[92m, \x1b[0m
var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// Regexes to extract fields
var (
    testNumRe   = regexp.MustCompile(`\btest=(\d+)\b`)
    containerRe = regexp.MustCompile(`\bcontainer=([a-z0-9]+)\b`)
    // Example lines:
    //   "[  10/2013] PASSED tests/..."
    //   "[gw3] [  10/2013] FAILED tests/..."
    simLineRe = regexp.MustCompile(`^(?:\[gw[^\]]+\]\s+)?\[\s*(\d+)\s*/\s*\d+\]\s+(PASSED|FAILED)\s+(.*)$`)
)

// Match EnqueueTransactions error arrays anywhere in a line, case-insensitive,
// and tolerant to minor variations like "error:" vs "err:".
var enqueueErrRe = regexp.MustCompile(`(?i)EnqueueTransactions\b[^\[]*?err(?:or)?:\s*(\[[^\]]*\])`)

// Run executes the analysis.
// - Reads test log, optionally prints cleaned lines, and maps test->container.
// - Parses simulator log into test entries.
// - Resolves client logs dir, extracts error arrays for failed tests, and writes summary.
func Run(testLogPath, simLogPath, clientLogsDirOverride, outPath string, printClean bool) error {
    // 1) Read and clean test.log lines, printing cleaned lines if requested.
    //    Collect a mapping testNumber -> containerID from lines containing both.
    testToContainer, err := mapTestsToContainers(testLogPath, printClean)
    if err != nil {
        return fmt.Errorf("reading test log: %w", err)
    }

    // 2) Parse simulator log for test number, status, and name.
    if simLogPath == "" {
        return errors.New("simulator log not found; specify with -simlog")
    }
    entries, err := parseSimulatorLog(simLogPath)
    if err != nil {
        return fmt.Errorf("reading simulator log: %w", err)
    }

    // 3) Determine client logs directory.
    clientLogsDir := clientLogsDirOverride
    if clientLogsDir == "" {
        clientLogsDir = resolveClientLogsDir(testLogPath, simLogPath)
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

    w := bufio.NewWriter(out)

    // cache for cid->error array to avoid re-reading files repeatedly
    cidErrCache := map[string]string{}
    // error occurrence counter across tests (each test contributes at most once per error)
    errCounts := map[string]int{}
    for _, e := range entries {
        cid := testToContainer[e.test]
        if e.status == "FAILED" {
            arr := lookupClientErrorArray(clientLogsDir, cid, cidErrCache)
            // Count errors once per test using minimal repeating unit
            unit := errorUnit(arr)
            // Do not count empty arrays in the summary to avoid noise,
            // but keep them in per-test output.
            if unit != "" && unit != "[]" {
                errCounts[unit]++
            }
            // Format: <test> <status> <container> <array> <test-name>
            fmt.Fprintf(w, "%s %s %s %s %s\n", e.test, e.status, cid, arr, e.name)
            continue
        }
        // PASSED: <test> <status> <container> <test-name>
        fmt.Fprintf(w, "%s %s %s %s\n", e.test, e.status, cid, e.name)
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

// mapTestsToContainers reads the test log, optionally prints cleaned lines to stdout,
// and returns a mapping from test number to container ID.
func mapTestsToContainers(testLogPath string, printClean bool) (map[string]string, error) {
    testToContainer := make(map[string]string)
    err := withFile(testLogPath, func(f *os.File) error {
        scanner := bufio.NewScanner(f)
        buf := make([]byte, scannerStartBuf)
        scanner.Buffer(buf, scannerMaxBuf)
        for scanner.Scan() {
            raw := scanner.Text()
            clean := ansiRe.ReplaceAllString(raw, "")
            if printClean {
                fmt.Println(clean)
            }
            if tn := firstGroup(testNumRe, clean); tn != "" {
                if cid := firstGroup(containerRe, clean); cid != "" {
                    testToContainer[tn] = cid
                }
            }
        }
        return scanner.Err()
    })
    if err != nil {
        return nil, err
    }
    return testToContainer, nil
}

// parseSimulatorLog extracts (test,status,name) tuples from the simulator log.
func parseSimulatorLog(simLogPath string) ([]simEntry, error) {
    entries := make([]simEntry, 0, 128)
    err := withFile(simLogPath, func(f *os.File) error {
        scanner := bufio.NewScanner(f)
        buf := make([]byte, scannerStartBuf)
        scanner.Buffer(buf, scannerMaxBuf)
        for scanner.Scan() {
            line := scanner.Text()
            if m := simLineRe.FindStringSubmatch(line); m != nil {
                entries = append(entries, simEntry{test: m[1], status: m[2], name: m[3]})
            }
        }
        return scanner.Err()
    })
    return entries, err
}

// resolveClientLogsDir attempts to find the wasp-client logs directory using
// common locations relative to the provided logs.
func resolveClientLogsDir(testLogPath, simLogPath string) string {
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

type simEntry struct {
    test   string
    status string // PASSED or FAILED
    name   string
}

func withFile(path string, fn func(*os.File) error) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    return fn(f)
}

func firstGroup(re *regexp.Regexp, s string) string {
    m := re.FindStringSubmatch(s)
    if len(m) > 1 {
        return m[1]
    }
    return ""
}

func lookupClientErrorArray(clientLogsDir, cid string, cache map[string]string) string {
    if cid == "" {
        return "[]"
    }
    if v, ok := cache[cid]; ok {
        return v
    }
    // find client log by prefix
    var matchPath string
    prefix := "client-" + cid
    // try fast glob first
    candidates, _ := filepath.Glob(filepath.Join(clientLogsDir, prefix+"*"))
    if len(candidates) > 0 {
        matchPath = candidates[0]
    } else {
        // fallback: scan dir (in case glob fails due to escaping)
        _ = filepath.WalkDir(clientLogsDir, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return nil
            }
            if d.IsDir() {
                return nil
            }
            name := d.Name()
            if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
                matchPath = path
                return fs.SkipDir // stop walking
            }
            return nil
        })
    }
    if matchPath == "" {
        cache[cid] = "[]"
        return "[]"
    }
    // scan file for first matching line
    var arr string = "[]"
    _ = withFile(matchPath, func(f *os.File) error {
        s := bufio.NewScanner(f)
        buf := make([]byte, scannerStartBuf)
        s.Buffer(buf, scannerMaxBuf)
        for s.Scan() {
            line := s.Text()
            // strip any ANSI codes that may be present in client logs
            clean := ansiRe.ReplaceAllString(line, "")
            if m := enqueueErrRe.FindStringSubmatch(clean); m != nil {
                arr = m[1]
                break
            }
        }
        return s.Err()
    })
    cache[cid] = arr
    return arr
}

// errorUnit normalizes an error slice string like "[msg msg]" or "[msg]" and
// returns a representative unit such that repeated identical messages within
// a test are collapsed to a single unit string.
func errorUnit(arr string) string {
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

