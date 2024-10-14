// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::anchor {
    use iota::{
        borrow::{Self, Referent, Borrow},
        coin::Coin,
        iota::IOTA,
    };
    use isc::{
        request::{Self, Request},
        assets_bag::{Self, AssetsBag},
    };

    // === Main structs ===

    /// An object which allows managing assets within the "ISC" ecosystem.
    /// By default it is owned by a single address.
    public struct Anchor has key, store {
        id: UID,
        assets: Referent<AssetsBag>,
        state_metadata: vector<u8>,
        state_index: u32,
    }

    public struct Receipt {
        /// ID of the request object
        request_id: ID,
    }

    // === Anchor packing and unpacking ===

    /// Starts a new chain by creating a new `Anchor` for it
    public fun start_new_chain(state_metadata: vector<u8>, coin: Option<Coin<IOTA>>, ctx: &mut TxContext): Anchor {
        let mut assets_bag = assets_bag::new(ctx);
        if (coin.is_some()) {
            assets_bag.place_coin<IOTA>(coin.destroy_some());
        } else {
            coin.destroy_none()
        };
        Anchor{
            id: object::new(ctx),
            assets: borrow::new(assets_bag, ctx),
            state_metadata: state_metadata,
            state_index: 0,
        }
    }

    /// Destroys an Anchor object and returns its assets bag.
    public fun destroy(self: Anchor): AssetsBag {
        let Anchor { id, assets, state_index: _, state_metadata: _ } = self;
        id.delete();
        assets.destroy()
    }

    // === Borrow assets from the Anchor ===

    /// Simulates a borrow mutable for the AssetsBag implementing the HotPotato pattern.
    public fun borrow_assets(self: &mut Anchor): (AssetsBag, Borrow) {
        borrow::borrow(&mut self.assets)
    }

    /// Finishes the simulation of a borrow mutable putting back the HotPotato.
    public fun return_assets_from_borrow(self: &mut Anchor, assets: AssetsBag, b: Borrow) {
        borrow::put_back(&mut self.assets, assets, b)
    }

    // === Receive a Request ===

    /// The Anchor receives a request and destroys it, implementing the HotPotato pattern.
    public fun receive_request(self: &mut Anchor, request: transfer::Receiving<Request>): (Receipt, AssetsBag) {
        let req = request::receive(&mut self.id, request);
        let (request_id, assets) = req.destroy();
        (Receipt { request_id }, assets)
    }

    public fun transition(self: &mut Anchor, new_state_metadata: vector<u8>, mut receipts: vector<Receipt>) {
        let receipts_len = receipts.length();
        let mut i = 0;
        while (i < receipts_len) {
            let Receipt {
                request_id: _
            } = receipts.pop_back(); // TODO it seems we don't need to use mut for `receipts`
            // here performs some cryptographic proof of inclusion with request_id and the new state root??
            i = i + 1;
        };
        receipts.destroy_empty();
        self.state_metadata = new_state_metadata;
    }

    // === Test Functions ===

    #[test_only]
    /// test only function to create a receipt
    public fun create_receipt_for_testing(request_id: ID): Receipt {
        Receipt { request_id }
    }
}
