// Package testdbhash provides utilities for calculating and verifying hash values of database contents
package testdbhash

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/samber/lo"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
)

const (
	envVarUpdateDBHash = "UPDATE_DBHASHES"

	// If set, a hex dump of the test will be stored in <basename>-<$DB_DUMP>.dump,
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
		false,
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
		true,
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
		true,
	)
}

func verifyHash(
	t solo.Context,
	iterateDBs func(func(chainID *isc.ChainID, k []byte, v []byte)),
	baseName string,
	msg string,
	isState bool,
) {
	h := lo.Must(blake2b.New256(nil))
	if h.Size() != hashing.HashSize {
		panic("unexpected h size")
	}

	var dbDump *os.File
	if os.Getenv(envVarDBDump) != "" {
		dumpFilename := baseName + "-" + os.Getenv(envVarDBDump) + ".dump"
		dbDump = lo.Must(os.Create(fullPath(dumpFilename)))
		defer dbDump.Close()
	}

	iterateDBs(func(_ *isc.ChainID, k []byte, v []byte) {
		lo.Must(h.Write(k))
		lo.Must(h.Write(v))
		if dbDump != nil {
			lo.Must(fmt.Fprintf(dbDump, "%s: [%d] %x\n", stringifyKey(k, isState), len(v), v))
		}
	})

	var hash hashing.HashValue
	copy(hash[:], h.Sum(nil))

	hashFilename := baseName + ".hex"
	if os.Getenv(envVarUpdateDBHash) != "" {
		saveHash(hashFilename, hash)
	} else {
		// TODO: replace Logf with Fatalf after tests are deterministic
		expected, err := loadHash(hashFilename)
		if err != nil {
			t.Logf("could not load hash from %s: %v", hashFilename, err)
		}
		if expected != hash {
			t.Logf(
				msg+
					" This may be due to a BREAKING CHANGE; make sure that you add a migration "+
					"(if necessary), and then run all tests again with: %s=1 (e.g. `%s=1 make test`). "+
					"Note: you can set %s=1 in one branch and %s=2 on another, and then compute a diff "+
					"of the generated hex dumps.",
				envVarUpdateDBHash, envVarUpdateDBHash, envVarDBDump, envVarDBDump,
			)
		}
	}
}

// stringifyKey formats the key in a human readable way, e.g. "[accounts|a]"
func stringifyKey(k []byte, isState bool) string {
	if !isState || len(k) < 4 {
		return hex.EncodeToString(k)
	}
	hname := codec.MustDecode[isc.Hname](k[:4])
	c, isCore := corecontracts.All[hname]
	if !isCore {
		return hex.EncodeToString(k)
	}
	// 99% of keys in the state have 1+ ASCII prefix
	rest := k[4:]
	asciiPrefix := rest[:0]
	for i := 0; i < len(rest) && rest[i] < unicode.MaxASCII && unicode.IsPrint(rune(rest[i])); i++ {
		asciiPrefix = rest[:i+1]
	}
	return fmt.Sprintf("[%s|%s] [%d] %x", c.Name, asciiPrefix, len(k), k)
}

func loadHash(filename string) (hashing.HashValue, error) {
	b, err := os.ReadFile(fullPath(filename))
	if err != nil {
		return hashing.HashValue{}, err
	}
	return hashing.MustHashValueFromHex(strings.TrimSpace(string(b))), nil
}

func saveHash(filename string, hash hashing.HashValue) {
	lo.Must0(os.WriteFile(fullPath(filename), []byte(hash.String()+"\n"), 0o600))
}

func fullPath(filename string) string {
	_, goFilename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(goFilename), normalize(filename))
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", "-"), "/", "-")
}
