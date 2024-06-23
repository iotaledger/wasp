// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::anchor {
    use sui::borrow::{Self, Referent, Borrow};   
    use isc::{
        request::{Self, Request},
        assets_bag::{Self, AssetsBag},
    }; 
       use sui::{
        table::{Self},
        
        coin::{Self, Coin},
        sui::SUI,
        url::{Self, Url},
        vec_set::{Self},
    };
    use std::string;

       use std::option;
    use sui::coin::{TreasuryCap};
    use sui::transfer;
    use sui::tx_context::{Self, TxContext};

    public struct TEST_A has drop {}
    public struct TEST_B has drop {}
    // === Main structs ===

    /// An object which allows managing assets within the "ISC" ecosystem.
    /// By default it is owned by a single address.
    public struct Anchor has key, store {
        id: UID,
        /// Anchor assets.
        assets: Referent<AssetsBag>,
        /// Anchor assets.
        state_root: vector<u8>,
        state_index: u32,
    }

        /// Make sure that the name of the type matches the module's name.
    public struct ANCHOR has drop {}

    public struct Receipt {
        /// ID of the request object
        request_id: ID,
    }

    public struct TestnetNFT has key, store {
    id: UID,
    name: string::String,
    description: string::String,
    url: url::Url
}

    // === Anchor packing and unpacking ===
    /// Module initializer is called once on module publish. A treasury
    /// cap is sent to the publisher, who then controls minting and burning
    fun init(witness: ANCHOR, ctx: &mut TxContext) {
        let (treasury, metadata) = coin::create_currency(witness, 6, b"MYCOIN", b"", b"", option::none(), ctx);
        
        transfer::public_freeze_object(metadata);
        transfer::public_transfer(treasury, tx_context::sender(ctx));
    }

    /// Starts a new chain by creating a new `Anchor` for it
    public fun start_new_chain(treasury_cap: &mut TreasuryCap<ANCHOR>,  ctx: &mut TxContext): Anchor {
        let coin = coin::mint(treasury_cap, 2000, ctx);
        let coin2 = coin::mint(treasury_cap, 2000, ctx);

        let mut assetsBag = assets_bag::new(ctx);
       assetsBag.place_coin(coin);
        assetsBag.place_coin(coin2);

let nft = TestnetNFT {
        id: object::new(ctx),
        name: string::utf8(vector::empty()),
        description: string::utf8(vector::empty()),
        url: url::new_unsafe_from_bytes(vector::empty())
    };

               assetsBag.place_asset(nft);

        let k = Anchor{
            id: object::new(ctx),
            assets: borrow::new(assetsBag, ctx),
            state_root: vector::empty(),
            state_index: 0,
         };

        k
    }

    /// Destroys an Anchor object and returns its assets bag.   
    public fun destroy(self: Anchor): AssetsBag {
        let Anchor { id, assets, state_root: _, state_index: _ } = self;
        id.delete();

        assets.destroy()
    }
    
    // === Borrow assets from the Anchor ===

    /// Simulates a borrow mutable for the AssetsBag implementing the HotPotato pattern.   
    public fun borrow_assets(self: &mut Anchor): (AssetsBag, Borrow) {
        borrow::borrow(&mut self.assets)
    }

    /// Finishes the simulation of a borrow mutable putting back the HotPotato. 
    public fun return_assets_from_borrow(
        self: &mut Anchor,
        assets: AssetsBag,
        b: Borrow
    ) {
        borrow::put_back(&mut self.assets, assets, b)
    }

    // === Receive a Request ===

    /// The Anchor receives a request and destroys it, implementing the HotPotato pattern.
    public fun receive_request(self: &mut Anchor, request: transfer::Receiving<Request>): (Receipt, AssetsBag) {
        let req = request::receive(&mut self.id, request);
        let (request_id, assets) = req.destroy();
        (Receipt { request_id }, assets)
    }

    public fun update_state_root(self: &mut Anchor, new_state_root: vector<u8>, mut receipts: vector<Receipt>) {
        let receipts_len = receipts.length();
        let mut i = 0;
        while (i < receipts_len) {
            let Receipt { 
                request_id: _ 
            } = receipts.pop_back();
            // here performs some cryptographic proof of inclusion with request_id and the new state root?? 
            i = i + 1;
        };
        receipts.destroy_empty();
        self.state_root = new_state_root
    } 
    

    // === Test Functions ===

    // test only function to create a receipt
    #[test_only]
    public fun create_receipt_for_testing(request_id: ID): Receipt {
        Receipt { request_id }
    }

}