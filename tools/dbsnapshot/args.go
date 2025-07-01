package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type BlockSelection struct {
	Blocks []int
}

func (bs *BlockSelection) String() string {
	return fmt.Sprintf("%v", bs.Blocks)
}

func (bs *BlockSelection) Set(value string) error {
	bs.Blocks = nil // Reset blocks

	if strings.Contains(value, "-") {
		// Handle range: "100-200"
		parts := strings.Split(value, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid range format, use: start-end")
		}

		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return fmt.Errorf("invalid start block: %v", err)
		}

		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return fmt.Errorf("invalid end block: %v", err)
		}

		if start > end {
			return fmt.Errorf("start block (%d) cannot be greater than end block (%d)", start, end)
		}

		for i := start; i <= end; i++ {
			bs.Blocks = append(bs.Blocks, i)
		}
	} else if strings.Contains(value, ",") {
		// Handle multiple blocks: "15,290,1000"
		parts := strings.Split(value, ",")
		for _, part := range parts {
			block, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return fmt.Errorf("invalid block number '%s': %v", part, err)
			}
			bs.Blocks = append(bs.Blocks, block)
		}
	} else {
		// Handle single block: "123"
		block, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("invalid block number: %v", err)
		}
		bs.Blocks = append(bs.Blocks, block)
	}

	bs.Blocks = removeDuplicates(bs.Blocks)
	sort.Ints(bs.Blocks)

	return nil
}

func removeDuplicates(slice []int) []int {
	keys := make(map[int]bool)
	var result []int

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func readArgs() (string, string, *BlockSelection) {
	var (
		blocks            = &BlockSelection{}
		sourceDBPath      = flag.String("source", "", "Source database path (required)")
		destinationDBPath = flag.String("destination", "", "Destination database path (required)")
	)

	flag.Var(blocks, "blocks", "Blocks to extract. Examples:\n"+
		"  Single block: -blocks 123\n"+
		"  Range: -blocks 100-200\n"+
		"  Multiple: -blocks 15,290,1000")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Block Extractor Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -source /path/to/source.db -target /path/to/target.db -blocks 123\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -source /path/to/source.db -target /path/to/target.db -blocks 100-200\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -source /path/to/source.db -target /path/to/target.db -blocks 15,290,1000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -source /path/to/source.db -target /path/to/target.db -blocks 1-10,50,100-110 -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validation
	if *sourceDBPath == "" {
		fmt.Fprintf(os.Stderr, "Error: source database path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *destinationDBPath == "" {
		fmt.Fprintf(os.Stderr, "Error: target database path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(blocks.Blocks) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no blocks specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	return *sourceDBPath, *destinationDBPath, blocks
}
