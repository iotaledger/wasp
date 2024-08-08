// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::request {
    use sui::{
        borrow::{Self, Referent},
        event::Self,
    };
    use isc::assets_bag::AssetsBag;
    use isc::allowance::Allowance;

    // === Main structs ===

    /// Contains the target contract, entry point and arguments
    public struct Message has drop, store {
        /// Contract name
        contract: u32,
        /// Function name
        function: u32,
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
        /// The target contract, entry point and arguments
        message: Message,
        /// The gas_budget of the request on L2
        allowance: Referent<Allowance>,
        /// The gas_budget of the request on L2
        gas_budget: u64,
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
        contract: u32,
        function: u32,
        args: vector<vector<u8>>,
        allowance: Allowance,
        gas_budget: u64,
        ctx: &mut TxContext,
    ) {
        send(Request{
            id: object::new(ctx),
            sender: ctx.sender(),
            assets_bag: borrow::new(assets_bag, ctx),
            message: Message{
                contract,
                function,
                args,
            },
            allowance: borrow::new(allowance, ctx),
            gas_budget: gas_budget,
        }, anchor)
    }

    /// Destroys a Request object and returns its balance and assets bag.
    public fun destroy(self: Request): (ID, AssetsBag, Allowance) {
        let Request {
            id,
            sender: _,
            assets_bag,
            message: _,
            allowance,
            gas_budget: _,
        } = self;
        let inner_id = id.uid_to_inner();
        id.delete();
        (inner_id, assets_bag.destroy(), allowance.destroy())
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

    #[test_only]
    /// test only function to create a request
    public fun create_for_testing(
        assets_bag: AssetsBag,
        contract: u32,
        function: u32,
        args: vector<vector<u8>>,
        allowance: Allowance,
        gas_budget: u64,
        ctx: &mut TxContext,
    ): Request {
        Request{
            id: object::new(ctx),
            sender: ctx.sender(),
            assets_bag: borrow::new(assets_bag, ctx),
            message: Message{
                contract,
                function,
                args,
            },
            allowance: borrow::new(allowance, ctx),
            gas_budget: gas_budget,
        }
    }
}
