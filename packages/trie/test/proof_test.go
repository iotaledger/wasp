package test

import (
	"io"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/trie"
)

func TestProofScenariosBlake2b(t *testing.T) {
	runScenario := func(name string, scenario []string) {
		t.Run(name, func(t *testing.T) {
			store := NewInMemoryKVStore()
			initRoot := lo.Must(trie.InitRoot(store, true))

			checklist, roots := runUpdateScenario(store, initRoot, scenario)
			trie.DebugDump(store, append([]trie.Hash{initRoot}, roots...), io.Discard)

			root := roots[len(roots)-1]
			trr := trie.NewReader(store, root)
			for k, v := range checklist {
				vBin := trr.Get([]byte(k))
				if v == "" {
					require.EqualValues(t, 0, len(vBin))
				} else {
					require.EqualValues(t, []byte(v), vBin)
				}
				p := trr.MerkleProof([]byte(k))
				err := p.Validate(root.Bytes())
				require.NoError(t, err)
				if v != "" {
					cID := trie.CommitToData([]byte(v))
					err = p.ValidateWithTerminal(root.Bytes(), cID.Bytes())
					require.NoError(t, err)
				} else {
					require.True(t, p.IsProofOfAbsence())
				}
			}
		})
	}
	runScenario("1", []string{"a"})
	runScenario("2", []string{"a", "ab"})
	runScenario("3", []string{"a", "ab", "a/"})
	runScenario("4", []string{"a", "ab", "a/", "ab/"})
	runScenario("5", []string{"a", "ab", "abc", "a/", "ab/"})
	runScenario("rnd", genRnd3())

	longData := make([]string, 0)
	for _, k := range []string{"a", "ab", "abc", "bca"} {
		longData = append(longData, k+"/"+strings.Repeat(k, 200))
	}
	runScenario("long", longData)
}
