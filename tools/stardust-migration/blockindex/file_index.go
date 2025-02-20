package blockindex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	"github.com/samber/lo"
)

func ReadIndexFromFile(filepath string) (_ []old_trie.Hash, fileExists bool) {
	cli.Logf("Reading index from %v\n", filepath)

	f, err := os.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false
		}

		panic(err)
	}

	defer f.Close()
	r := bufio.NewReader(f)

	fSize := lo.Must(f.Stat()).Size()
	indexEntries := fSize / int64(old_trie.HashSizeBytes)
	res := make([]old_trie.Hash, 0, indexEntries)

	for {
		var trieRootBytes old_trie.Hash
		if n, err := io.ReadFull(r, trieRootBytes[:]); err != nil {
			if !errors.Is(err, io.EOF) {
				panic(err)
			}

			if n != 0 {
				panic("unexpected number of bytes left at the end of the file")
			}

			break
		} else if n != len(trieRootBytes) {
			panic(fmt.Sprintf("unexpected number of bytes when reading entry [%v]: expected %v, got %v", len(res), len(trieRootBytes), n))
		}

		res = append(res, trieRootBytes)
	}

	cli.Logf("Found %v index entries", len(res))

	return res, true
}
