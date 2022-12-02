// package trie implements an immutable Merkle Patricia Trie,
// used by the state package to store the chain state.
//
// The code is based on the trie.go package by Evaldas Drasutis:
// https://github.com/iotaledger/trie.go
//
// This is a simplified version, keeping only the features that are relevant
// for our use case. Namely:
//
// - Arity is fixed at 16
// - Hash size is fixed at 20 bytes
// - Hashing algorithm is blake2b
// - No mutable trie implementation
package trie
