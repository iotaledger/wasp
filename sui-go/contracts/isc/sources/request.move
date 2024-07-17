// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::request {
    use std::string::String;
    use sui::{
        borrow::{Self, Referent},
        event::Self,
    };
    use isc::assets_bag::AssetsBag;

    // === Main structs ===

    /// Contains the request data to be used off-chain.
    public struct RequestData has drop, store {
        /// Contract name
        contract: String,
        /// Function name
        function: String,
        /// Function arguments
        args: vector<vector<u8>>,
    }

    /// Represents a request object
    public struct Request has key {
        id: UID,
        /// Request sender
        sender: address,
        /// Bag of assets associated to the request
        assets_bag: Referent<AssetsBag>,
        /// The request data, to be used off-chain 
        data: Option<RequestData>,
    }

    // === Events ===

    /// Emitted when a request is sent to an address.
    public struct RequestEvent has copy, drop {
        /// ID of the request object
        request_id: ID,
        /// Anchor receiving the request
        anchor: address,
    }

    // === Request packing and unpacking ===

    /// Creates a request to call a specific SC function.
    public fun create_and_send_request(
        anchor: address,
        assets_bag: AssetsBag,
        contract: Option<String>,
        function: Option<String>,
        args: Option<vector<vector<u8>>>,
        ctx: &mut TxContext,
    ) {
        let id = object::new(ctx);
        let data = if (option::is_some(&contract) 
            && option::is_some(&function)
            && option::is_some(&args)) {
                option::some(RequestData {
                    contract: contract.destroy_some(),
                    function: function.destroy_some(),
                    args: args.destroy_some(),
                })
            } else {
                option::none()
            };
        send(Request{
            id,
            sender: ctx.sender(),
            assets_bag: borrow::new(assets_bag, ctx),
            data,
        }, anchor)
    }

    /// Destroys a Request object and returns its balance and assets bag.
    public fun destroy(self: Request): (ID, AssetsBag) {
        let Request {
            id,
            sender: _,
            assets_bag,
            data: _,
        } = self;
        let inner_id = id.uid_to_inner();
        id.delete();
        (inner_id, assets_bag.destroy())
    }

    // === Send and receive the Request ===

    /// Send a Request object to an anchor and emits the RequestEvent.
    fun send(self: Request, anchor: address) {
        event::emit(RequestEvent { request_id: self.id.uid_to_inner(), anchor });
        transfer::transfer(self, anchor)
    }

    /// Utility function to receive a `Request` object in other ISC modules.
    /// Other modules in the ISC package can call this function to receive an `Request` object.
    public(package) fun receive(parent: &mut UID, self: transfer::Receiving<Request>): Request {
        transfer::receive(parent, self)
    }

    // === Test Functions ===

    /// test only function to create a request
    #[test_only]
    public fun create_for_testing(
        assets_bag: AssetsBag,
        contract: Option<String>,
        function: Option<String>,
        args: Option<vector<vector<u8>>>,
        ctx: &mut TxContext,
    ): Request {
        let id = object::new(ctx);
        let data = if (option::is_some(&contract) 
            && option::is_some(&function)
            && option::is_some(&args)) {
                option::some(RequestData {
                    contract: contract.destroy_some(),
                    function: function.destroy_some(),
                    args: args.destroy_some(),
                })
            } else {
                option::none()
            };
        Request{
            id,
            sender: ctx.sender(),
            assets_bag: borrow::new(assets_bag, ctx),
            data,
        }
    }
}