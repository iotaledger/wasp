// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crypto::signatures::ed25519;

pub struct KeyPair {
    private_key: ed25519::SecretKey,
    public_key: ed25519::PublicKey,
}
