// Package readonly provides functionality for components read/check whether readobly mode is set by CLI
package readonly

import "path/filepath"

// Enabled returns true when a read-only database root has been provided.
func Enabled(root string) bool { return root != "" }

// DataDir returns the path to the chain state data directory within the
// provided read-only database root.
func DataDir(root string) string { return filepath.Join(root, "data") }

// ChainRegistryFile returns the path to the chain registry file within the
// provided read-only database root.
func ChainRegistryFile(root string) string { return filepath.Join(root, "chain_registry.json") }
