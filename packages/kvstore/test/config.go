//go:build !rocksdb
// +build !rocksdb

package kvstore_test //nolint:staticcheck

var dbImplementations = []string{"mapDB"}
