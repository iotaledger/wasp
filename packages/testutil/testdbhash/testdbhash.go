package testdbhash

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/solo"
)

const envvar = "UPDATE_DBHASHES"

func VerifyDBHash(env *solo.Solo, testName string) {
	if os.Getenv(envvar) != "" {
		saveHash(testName, env.GetDBHash())
		return
	}
	expected := loadHash(testName)
	actual := env.GetDBHash()
	if expected != actual {
		env.T.Fatalf(
			"DB hash has changed! "+
				"This is a BREAKING CHANGE; make sure that you add a migration "+
				"(if necessary), and then run all tests again with: %s=1 (e.g. `%s=1 make test`)",
			envvar, envvar,
		)
	}
}

func loadHash(testName string) hashing.HashValue {
	b, err := os.ReadFile(hashFileName(testName))
	if err != nil {
		panic(err)
	}
	return hashing.MustHashValueFromHex(strings.TrimSpace(string(b)))
}

func saveHash(testName string, hash hashing.HashValue) {
	s := hash.String()
	err := os.WriteFile(hashFileName(testName), []byte(s+"\n"), 0o600)
	if err != nil {
		panic(err)
	}
}

func hashFileName(testName string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), testName+".hex")
}
