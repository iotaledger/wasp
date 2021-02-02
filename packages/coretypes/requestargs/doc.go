// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package encodedargs implements special encoding of the dict.Dict which alows
// optimized transfer of big data through SC request. It encodes big data chunks
// as hashes, whihc later can be decoded (solidified) into the original form
package requestargs
