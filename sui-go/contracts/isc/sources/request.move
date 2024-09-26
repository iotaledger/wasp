// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::request {
    use sui::{
        borrow::{Self, Referent},
        event::Self,
    };
    use std::{
        string::String,
    };
    use isc::assets_bag::AssetsBag;

    // The allowance coin_types vector and balances vector are not in the same size
    const EAllowanceVecUnequal: u64 = 0;

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

    public struct CoinAllowance has drop, store {
        coin_type: String,
        balance: u64,
    }

    /// Represents a request object
    public struct Request has key {
        id: UID,
        /// Request sender
        sender: address,
        /// Contract identity of the sender contract.
        /// It is set when a contract calls cross chain request.
        contract_identity: vector<u8>,
        /// Bag of assets associated to the request
        assets_bag: Referent<AssetsBag>,
        /// The target contract, entry point and arguments
        message: Message,
        /// The gas_budget of the request on L2
        allowance: vector<CoinAllowance>,
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
        contract_identity: Option<vector<u8>>,
        contract: u32,
        function: u32,
        args: vector<vector<u8>>,
        mut allowance_cointypes: vector<String>,
        mut allowance_balances: vector<u64>,
        gas_budget: u64,
        ctx: &mut TxContext,
    ) {
        let mut allowance_cointypes_len = vector::length<String>(&allowance_cointypes);
        assert!(allowance_cointypes_len == vector::length<u64>(&allowance_balances), EAllowanceVecUnequal);

        let mut allowance = vector::empty();
        while (allowance_cointypes_len > 0) {
            let allowance_elt = CoinAllowance {
                coin_type: allowance_cointypes.pop_back(),
                balance: allowance_balances.pop_back(),
            };
            allowance.push_back(allowance_elt);
            allowance_cointypes_len = allowance_cointypes_len - 1;
        };

        let contract_identity_extract = if (contract_identity.is_some()) {
            option::destroy_some<vector<u8>>(contract_identity)
        } else {
            vector::empty()
        };

        send(Request{
            id: object::new(ctx),
            sender: ctx.sender(),
            contract_identity: contract_identity_extract,
            assets_bag: borrow::new(assets_bag, ctx),
            message: Message{
                contract,
                function,
                args,
            },
            allowance: allowance,
            gas_budget: gas_budget,
        }, anchor)
    }

    /// Destroys a Request object and returns its balance and assets bag.
    public fun destroy(self: Request): (ID, AssetsBag) {
        let Request {
            id,
            sender: _,
            contract_identity: _,
            assets_bag,
            message: _,
            allowance: _,
            gas_budget: _,
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

    #[test_only]
    /// test only function to create a request
    public fun create_for_testing(
        assets_bag: AssetsBag,
        contract: u32,
        function: u32,
        args: vector<vector<u8>>,
        mut allowance_cointypes: vector<String>,
        mut allowance_balances: vector<u64>,
        gas_budget: u64,
        ctx: &mut TxContext,
    ): Request {
        let mut allowance_cointypes_len = vector::length<String>(&allowance_cointypes);
        assert!(allowance_cointypes_len == vector::length<u64>(&allowance_balances), EAllowanceVecUnequal);

        let mut allowance = vector::empty();
        while (allowance_cointypes_len > 0) {
            let allowance_elt = CoinAllowance {
                coin_type: allowance_cointypes.pop_back(),
                balance: allowance_balances.pop_back(),
            };
            allowance.push_back(allowance_elt);
            allowance_cointypes_len = allowance_cointypes_len - 1;
        };

        Request{
            id: object::new(ctx),
            sender: ctx.sender(),
            assets_bag: borrow::new(assets_bag, ctx),
            message: Message{
                contract,
                function,
                args,
            },
            allowance: allowance,
            gas_budget: gas_budget,
        }
    }
}
