package trie

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProofScenariosBlake2b(t *testing.T) {
	runScenario := func(name string, scenario []string) {
		t.Run(name, func(t *testing.T) {
			store := NewInMemoryKVStore()
			initRoot := MustInitRoot(store)
			tr, err := NewTrieUpdatable(store, initRoot)
			require.NoError(t, err)

			checklist, root := runUpdateScenario(tr, store, scenario)
			trr, err := NewTrieReader(store, root)
			require.NoError(t, err)
			for k, v := range checklist {
				vBin := trr.Get([]byte(k))
				if v == "" {
					require.EqualValues(t, 0, len(vBin))
				} else {
					require.EqualValues(t, []byte(v), vBin)
				}
				p := trr.MerkleProof([]byte(k))
				err = p.Validate(root.Bytes())
				require.NoError(t, err)
				if len(v) > 0 {
					cID := commitToData([]byte(v))
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
