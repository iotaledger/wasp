// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package requestargs implements special encoding of the dict.Dict which allows
// optimized transfer of big data through SC request. It encodes big data chunks
// as hashes, which later can be decoded (solidified) into the original form treating the hashes
// as content addresses on IPFS and other mediums well suited for dalivery of big data chunks to wasm nodes
package requestargs
