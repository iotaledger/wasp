package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	storageFiles := []string{"storage_sol.json", "storage_rs.json", "storage_ts.json", "storage_go.json"}
	memoryFiles := []string{"memory_sol.json", "memory_rs.json", "memory_ts.json", "memory_go.json"}
	exetionTimeFiles := []string{"executiontime_sol.json", "executiontime_rs.json", "executiontime_ts.json", "executiontime_go.json"}

	drawGraph("Storage contract gas usage", "storage", storageFiles)
	drawGraph("Memory contract gas usage", "memory", memoryFiles)
	drawGraph("Execution time contract gas usage", "executiontime", exetionTimeFiles)
}

func drawGraph(title, contract string, filenames []string) {
	p := plot.New()

	p.Title.Text = title
	p.X.Label.Text = "N"
	p.Y.Label.Text = "Gas"

	v := make([]interface{}, 0)
	for _, filename := range filenames {
		filePath := path.Join("../", contract, "pkg", filename)
		bytes, err := os.ReadFile(filePath)
		check(err)

		var points map[uint32]uint64
		err = json.Unmarshal(bytes, &points)
		check(err)

		graphTitle, xys := graphTitle(filename), graphData(filePath, points)
		v = append(v, graphTitle, xys)
	}
	err := plotutil.AddLinePoints(p, v...)
	check(err)

	err = p.Save(8*vg.Inch, 8*vg.Inch, contract+".png")
	check(err)
}

func graphData(filename string, points map[uint32]uint64) plotter.XYs {
	_ = filename
	xys := make(plotter.XYs, 0)
	for x, y := range points {
		xys = append(xys, plotter.XY{X: float64(x), Y: float64(y)})
	}
	return xys
}

func graphTitle(filename string) string {
	if strings.Contains(filename, "go") {
		return "Golang"
	} else if strings.Contains(filename, "rs") {
		return "Rust"
	} else if strings.Contains(filename, "ts") {
		return "Typescript"
	}
	return "Solidity"
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
