package main

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var rootCmd *cobra.Command

func initRootCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Version: "0.1.0",
		Use:     "gascalibration",
		Args:    cobra.NoArgs,
		Short:   "gascalibration is a command line tool to generate gas calibration reports.",
		Long:    `gascalibration is a command line tool to generate gas calibration reports from storage, memory and execution time contracts.`,
		Run: func(cmd *cobra.Command, args []string) {
			storageFiles := []string{"storage_sol.json", "storage_rs.json", "storage_ts.json", "storage_go.json"}
			memoryFiles := []string{"memory_sol.json", "memory_rs.json", "memory_ts.json", "memory_go.json"}
			executionTimeFiles := []string{"executiontime_sol.json", "executiontime_rs.json", "executiontime_ts.json", "executiontime_go.json"}

			drawGraph := graphDrawer(dir)
			drawGraph("Storage contract gas usage", "storage", storageFiles)
			drawGraph("Memory contract gas usage", "memory", memoryFiles)
			drawGraph("Execution time contract gas usage", "executiontime", executionTimeFiles)
		},
	}

	cmd.PersistentFlags().StringVar(&dir, "dir", "", "Directory containing contracts")

	return cmd
}

func init() {
	rootCmd = initRootCmd()
}

func main() {
	log.Check(rootCmd.Execute())
}

func graphDrawer(dir string) func(string, string, []string) {
	return func(title, contract string, filenames []string) {
		p := plot.New()

		p.Title.Text = title
		p.X.Label.Text = "N"
		p.Y.Label.Text = "Gas"

		v := make([]interface{}, 0)
		for _, filename := range filenames {
			filePath := path.Join(dir, contract, "pkg", filename)
			bytes, err := os.ReadFile(filePath)
			log.Check(err)

			var points map[uint32]uint64
			err = json.Unmarshal(bytes, &points)
			log.Check(err)

			graphTitle, xys := graphTitle(filename), graphData(points)
			v = append(v, graphTitle, xys)
		}
		err := plotutil.AddLinePoints(p, v...)
		log.Check(err)

		filePath := path.Join(dir, contract+".png")
		err = p.Save(8*vg.Inch, 8*vg.Inch, filePath)
		log.Check(err)
	}
}

func graphData(points map[uint32]uint64) plotter.XYs {
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
