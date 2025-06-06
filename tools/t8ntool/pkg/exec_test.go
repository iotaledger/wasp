// Copyright 2020 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package t8ntool_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	t8ntool "github.com/iotaledger/wasp/tools/t8ntool/pkg"
	"github.com/iotaledger/wasp/tools/t8ntool/pkg/cmdtest"
	"github.com/iotaledger/wasp/tools/t8ntool/pkg/reexec"
	"github.com/urfave/cli/v2"
)

func TestMain(m *testing.M) {
	// Run the app if we've been exec'd as "ethkey-test" in runEthkey.
	app := &cli.App{
		Name:  "t8n",
		Usage: "a test main app",
		Commands: []*cli.Command{
			&cli.Command{
				Name:    "transition",
				Aliases: []string{"t8n"},
				Usage:   "Executes a full state transition",
				Action:  t8ntool.Transition,
				Flags: []cli.Flag{
					t8ntool.TraceFlag,
					t8ntool.TraceTracerFlag,
					t8ntool.TraceTracerConfigFlag,
					t8ntool.TraceEnableMemoryFlag,
					t8ntool.TraceDisableStackFlag,
					t8ntool.TraceEnableReturnDataFlag,
					t8ntool.TraceEnableCallFramesFlag,
					t8ntool.OutputBasedir,
					t8ntool.OutputAllocFlag,
					t8ntool.OutputResultFlag,
					t8ntool.OutputBodyFlag,
					t8ntool.InputAllocFlag,
					t8ntool.InputEnvFlag,
					t8ntool.InputTxsFlag,
					t8ntool.ForknameFlag,
					t8ntool.ChainIDFlag,
					t8ntool.RewardFlag,
				},
			},
		},
	}

	reexec.Register("t8n-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}

func TestT8n(t *testing.T) {
	t.Parallel()
	tt := new(testT8n)
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	for i, tc := range []struct {
		base        string
		input       t8nInput
		output      t8nOutput
		expExitCode int
		expOut      string
	}{
		{ // Test exit (3) on bad config
			base: "./testdata/1",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Frontier+1346", "",
			},
			output:      t8nOutput{alloc: true, result: true},
			expExitCode: 3,
		},
		{
			base: "./testdata/1",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Byzantium", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // blockhash test
			base: "./testdata/3",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Berlin", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // missing blockhash test
			base: "./testdata/4",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Berlin", "",
			},
			output:      t8nOutput{alloc: true, result: true},
			expExitCode: 4,
		},
		{ // Uncle test
			base: "./testdata/5",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Byzantium", "0x80",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // Sign json transactions
			base: "./testdata/13",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "London", "",
			},
			output: t8nOutput{body: true},
			expOut: "exp.json",
		},
		{ // Already signed transactions
			base: "./testdata/13",
			input: t8nInput{
				"alloc.json", "signed_txs.rlp", "env.json", "London", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp2.json",
		},
		{ // Difficulty calculation - no uncles
			base: "./testdata/14",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "London", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp.json",
		},
		{ // Difficulty calculation - with uncles
			base: "./testdata/14",
			input: t8nInput{
				"alloc.json", "txs.json", "env.uncles.json", "London", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp2.json",
		},
		{ // Difficulty calculation - with ommers + Berlin
			base: "./testdata/14",
			input: t8nInput{
				"alloc.json", "txs.json", "env.uncles.json", "Berlin", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp_berlin.json",
		},
		{ // Difficulty calculation on arrow glacier
			base: "./testdata/19",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "London", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp_london.json",
		},
		{ // Difficulty calculation on arrow glacier
			base: "./testdata/19",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "ArrowGlacier", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp_arrowglacier.json",
		},
		{ // Difficulty calculation on gray glacier
			base: "./testdata/19",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "GrayGlacier", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp_grayglacier.json",
		},
		{ // Sign unprotected (pre-EIP155) transaction
			base: "./testdata/23",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Berlin", "",
			},
			output: t8nOutput{result: true},
			expOut: "exp.json",
		},
		{ // Test post-merge transition
			base: "./testdata/24",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Paris", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // Test post-merge transition where input is missing random
			base: "./testdata/24",
			input: t8nInput{
				"alloc.json", "txs.json", "env-missingrandom.json", "Paris", "",
			},
			output:      t8nOutput{alloc: false, result: false},
			expExitCode: 3,
		},
		{ // Test base fee calculation
			base: "./testdata/25",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Paris", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // Test withdrawals transition
			base: "./testdata/26",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Shanghai", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // Cancun tests
			base: "./testdata/28",
			input: t8nInput{
				"alloc.json", "txs.rlp", "env.json", "Cancun", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // More cancun tests
			base: "./testdata/29",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Cancun", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // More cancun test, plus example of rlp-transaction that cannot be decoded properly
			base: "./testdata/30",
			input: t8nInput{
				"alloc.json", "txs_more.rlp", "env.json", "Cancun", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
		{ // Prague test, EIP-7702 transaction
			base: "./testdata/33",
			input: t8nInput{
				"alloc.json", "txs.json", "env.json", "Prague", "",
			},
			output: t8nOutput{alloc: true, result: true},
			expOut: "exp.json",
		},
	} {
		args := []string{"t8n"}
		args = append(args, tc.output.get()...)
		args = append(args, tc.input.get(tc.base)...)
		var qArgs []string // quoted args for debugging purposes
		for _, arg := range args {
			if len(arg) == 0 {
				qArgs = append(qArgs, `""`)
			} else {
				qArgs = append(qArgs, arg)
			}
		}
		tt.Logf("args: %v\n", strings.Join(qArgs, " "))
		tt.Run("t8n-test", args...)
		// Compare the expected output, if provided
		if tc.expOut != "" {
			file := fmt.Sprintf("%v/%v", tc.base, tc.expOut)
			want, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("test %d: could not read expected output: %v", i, err)
			}
			have := tt.Output()
			// fmt.Println("have: ", string(have))
			ok, err := cmpJson(have, want)
			switch {
			case err != nil:
				t.Fatalf("test %d, file %v: json parsing failed: %v", i, file, err)
			case !ok:
				t.Fatalf("test %d, file %v: output wrong, have \n%v\nwant\n%v\n", i, file, string(have), string(want))
			}
		}
		tt.WaitExit()
		if have, want := tt.ExitStatus(), tc.expExitCode; have != want {
			t.Fatalf("test %d: wrong exit code, have %d, want %d", i, have, want)
		}
	}
}

type testT8n struct {
	*cmdtest.TestCmd
}

type t8nInput struct {
	inAlloc  string
	inTxs    string
	inEnv    string
	stFork   string
	stReward string
}

func (args *t8nInput) get(base string) []string {
	var out []string
	if opt := args.inAlloc; opt != "" {
		out = append(out, "--input.alloc")
		out = append(out, fmt.Sprintf("%v/%v", base, opt))
	}
	if opt := args.inTxs; opt != "" {
		out = append(out, "--input.txs")
		out = append(out, fmt.Sprintf("%v/%v", base, opt))
	}
	if opt := args.inEnv; opt != "" {
		out = append(out, "--input.env")
		out = append(out, fmt.Sprintf("%v/%v", base, opt))
	}
	if opt := args.stFork; opt != "" {
		out = append(out, "--state.fork", opt)
	}
	if opt := args.stReward; opt != "" {
		out = append(out, "--state.reward", opt)
	}
	return out
}

type t8nOutput struct {
	alloc  bool
	result bool
	body   bool
}

func (args *t8nOutput) get() (out []string) {
	if args.body {
		out = append(out, "--output.body", "stdout")
	} else {
		out = append(out, "--output.body", "") // empty means ignore
	}
	if args.result {
		out = append(out, "--output.result", "stdout")
	} else {
		out = append(out, "--output.result", "")
	}
	if args.alloc {
		out = append(out, "--output.alloc", "stdout")
	} else {
		out = append(out, "--output.alloc", "")
	}
	return out
}

func cmpJson(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}
