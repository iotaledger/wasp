package gascalibration

import (
	"encoding/json"
	"log"
	"os"
)

func SaveTestResultAsJSON(filepath string, results map[uint32]uint64) {
	f, err := os.Create(filepath)
	check(err)
	defer f.Close()

	bytes, err := json.Marshal(results)
	check(err)

	_, err = f.Write(bytes)
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
