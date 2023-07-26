package cli

import (
	"syscall"

	"github.com/awnumar/memguard"
	"golang.org/x/term"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func ReadPasswordFromStdin() *memguard.Enclave {
	log.Printf("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:nolintlint,unconvert // int cast is needed for windows
	log.Check(err)
	return memguard.NewEnclave(passwordBytes)
}
