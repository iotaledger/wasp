package blockindex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	"github.com/samber/lo"
)

func NewFileIndexer(filepath string, store old_state.Store) (_ *FileIndexer, fileExists bool) {
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
	blockTrieRoots := make([]old_trie.Hash, 0, indexEntries)

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
			panic(fmt.Sprintf("unexpected number of bytes when reading entry [%v]: expected %v, got %v", len(blockTrieRoots), len(trieRootBytes), n))
		}

		blockTrieRoots = append(blockTrieRoots, trieRootBytes)
	}

	cli.Logf("Found %v index entries", len(blockTrieRoots))

	latestBlockIndex := lo.Must(store.LatestBlockIndex())
	if len(blockTrieRoots) != int(latestBlockIndex+1) {
		cli.Logf("** WARNING: index file IGNORED - it was created for other database: block in db = %v, index entries = %v", len(blockTrieRoots), latestBlockIndex+1)
		return nil, false
	}

	return &FileIndexer{
		store:          store,
		blockTrieRoots: blockTrieRoots,
	}, true
}

type FileIndexer struct {
	store          old_state.Store
	blockTrieRoots []old_trie.Hash
}

var _ BlockIndex = &FileIndexer{}

func (ind *FileIndexer) BlockByIndex(i uint32) (old_state.Block, old_trie.Hash) {
	trieRoot := ind.blockTrieRoots[i]
	block := lo.Must(ind.store.BlockByTrieRoot(trieRoot))
	return block, trieRoot
}

func (ind *FileIndexer) BlocksCount() uint32 {
	return uint32(len(ind.blockTrieRoots))
}
