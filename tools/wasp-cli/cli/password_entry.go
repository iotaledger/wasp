// Package cli implements the command-line interface for the wasp-cli tool,
// providing core functionality and utilities for user interaction.
package cli

import (
	"syscall"

	"github.com/awnumar/memguard"
	"golang.org/x/term"

	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func ReadPasswordFromStdin() *memguard.Enclave {
	log.Printf("\nPassword required to open/create secured storage.\n")
	log.Printf("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:nolintlint,unconvert // int cast is needed for windows
	log.Check(err)
	log.Printf("\n")
	return memguard.NewEnclave(passwordBytes)
}
