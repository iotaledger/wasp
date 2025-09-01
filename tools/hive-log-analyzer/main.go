package main

import (
	"flag"
	"log"
	"path/filepath"
)

// This program is used for analyzig the result of hive tests
// running command like './hive --sim ethereum/eest --client wasp-client -docker.pull -sim.parallelism 10 > test.log 2>&1'
// and provide the path of each files as the following flags request
func main() {
	testLogPathFlag := flag.String("testlog", "./workspace/test.log", "Path to test.log with ANSI codes")
	simLogPathFlag := flag.String("simlog", "", "Path to simulator log (e.g. *-simulator-*.log)")
	clientLogsFlag := flag.String("clientlogs", "", "Optional path to wasp-client logs directory")
	printCleanFlag := flag.Bool("print-clean", true, "Print cleaned test.log lines to stdout")
	outPathFlag := flag.String("out", "./hive-summary.txt", "Output file with test results")
	flag.Parse()

	testLogPath := *testLogPathFlag
	simLogPath := *simLogPathFlag
	clientLogsPath := *clientLogsFlag
	outPath := *outPathFlag
	printClean := *printCleanFlag

	if simLogPath == "" {
		// Try to discover the simulator log next to the test log
		dir := filepath.Dir(testLogPath)
		matches, _ := filepath.Glob(filepath.Join(dir, "*-simulator-*.log"))
		if len(matches) > 0 {
			simLogPath = matches[0]
		}
	}

	if err := Run(testLogPath, simLogPath, clientLogsPath, outPath, printClean); err != nil {
		log.Fatal(err)
	}
}
