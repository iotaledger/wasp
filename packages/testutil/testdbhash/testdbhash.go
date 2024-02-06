package testdbhash

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/solo"
)

const (
	envVarUpdateDBHash = "UPDATE_DBHASHES"

	// If set, a hex dump of the test will be stored in <basename>-<$DB_DUMP>.dump.hex,
	// that can be used to compute a diff.
	envVarDBDump = "DB_DUMP"
)

// VerifyDBHash calculates a hash of the database contents, compares it to the hash stored in
// <testName>.hex, and panics if the hash changed.
// The DB hash includes the whole chain DB, and that includes the whole trie, all
// blocks, all states, etc, making it difficult to tell what caused the change.
func VerifyDBHash(env *solo.Solo, testName string) {
	verifyHash(
		env.T,
		env.IterateChainTrieDBs,
		testName,
		"DB hash has changed!",
	)
}

// VerifyStateHash calculates a hash of the chain state at the latest state index,
// compares it to the hash stored in <testName>-<contract>.hex, and panics if the hash changed.
func VerifyStateHash(env *solo.Solo, testName string) {
	verifyHash(
		env.T,
		func(f func(chainID *isc.ChainID, k []byte, v []byte)) {
			env.IterateChainLatestStates("", f)
		},
		testName+"-state",
		"State hash has changed!",
	)
}

// VerifyContractStateHash calculates a hash of the contract state at the latest state index,
// compares it to the hash stored in <testName>-<contract>.hex, and panics if the hash changed.
func VerifyContractStateHash(env *solo.Solo, contract *coreutil.ContractInfo, prefix kv.Key, testName string) {
	verifyHash(
		env.T,
		func(f func(chainID *isc.ChainID, k []byte, v []byte)) {
			env.IterateChainLatestStates(kv.Key(contract.Hname().Bytes())+prefix, f)
		},
		testName+"-"+contract.Name,
		fmt.Sprintf("State hash for core contract %q has changed!", contract.Name),
	)
}

func verifyHash(
	t solo.Context,
	iterateDBs func(func(chainID *isc.ChainID, k []byte, v []byte)),
	baseName string,
	msg string,
) {
	h := lo.Must(blake2b.New256(nil))
	if h.Size() != hashing.HashSize {
		panic("unexpected h size")
	}

	var dbDump *os.File
	if os.Getenv(envVarDBDump) != "" {
		dumpFilename := baseName + "-" + os.Getenv(envVarDBDump) + ".dump.hex"
		dbDump = lo.Must(os.Create(fullPath(dumpFilename)))
		defer dbDump.Close()
	}

	iterateDBs(func(_ *isc.ChainID, k []byte, v []byte) {
		lo.Must(h.Write(k))
		lo.Must(h.Write(v))
		if dbDump != nil {
			lo.Must(dbDump.WriteString(fmt.Sprintf("%x: %x\n", k, v)))
		}
	})

	var hash hashing.HashValue
	copy(hash[:], h.Sum(nil))

	hashFilename := baseName + ".hex"
	if os.Getenv(envVarUpdateDBHash) != "" {
		saveHash(hashFilename, hash)
	} else {
		expected := loadHash(hashFilename)
		if expected != hash {
			t.Fatalf(
				msg+
					" This may be a BREAKING CHANGE; make sure that you add a migration "+
					"(if necessary), and then run all tests again with: %s=1 (e.g. `%s=1 make test`). "+
					"Note: you can set %s=1 in one branch and %s=2 on another, and then compute a diff "+
					"of the generated hex dumps.",
				envVarUpdateDBHash, envVarUpdateDBHash, envVarDBDump, envVarDBDump,
			)
		}
	}
}

func loadHash(filename string) hashing.HashValue {
	b := lo.Must(os.ReadFile(fullPath(filename)))
	return hashing.MustHashValueFromHex(strings.TrimSpace(string(b)))
}

func saveHash(filename string, hash hashing.HashValue) {
	lo.Must0(os.WriteFile(fullPath(filename), []byte(hash.String()+"\n"), 0o600))
}

func fullPath(filename string) string {
	_, goFilename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(goFilename), filename)
}
