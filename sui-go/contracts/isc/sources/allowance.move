// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
module isc::allowance {
    use std::type_name;
    use sui::{
        balance::Balance,
        dynamic_field as df,
    };

    // Asset hasn't been added into Allowance
    const EAllowanceNotExist: u64 = 0;

    // === Main structs ===

    /// An object contains the allowance of a request. All the allowance should be on the L2
    /// when execting the request.
    // So an allowance is only a map of values. The ownership of the referred assets has 
    // been moved L2 when execting the request.
    public struct Allowance has key, store {
        id: UID,
    }

    // === Dynamic Field keys ===

    // === Allowance packing and unpacking ===

    /// Creates a new empty `Allowance`
    public fun new(ctx: &mut TxContext): Allowance {
        Allowance{
            id: object::new(ctx),
        }
    }

    /// Destroys an Allowance object and returns its balance.
    public fun destroy(self: Allowance) {
        let Allowance { id } = self;
        id.delete();
    }

    // === Add values into the Allowance ===
    
    public fun add_coin_allowance<T>(self: &mut Allowance, allowance_balance: &Balance<T>) {
        let coin_type = type_name::get<T>().into_string();
        let balance_val = allowance_balance.value();
        if(df::exists_(&self.id, coin_type)) {
            let placed_val = df::borrow_mut<std::ascii::String, u64>(&mut self.id, coin_type);
            *placed_val = *placed_val + balance_val;
        } else {
            df::add(&mut self.id, coin_type, balance_val);
        }
    }

    public fun add_object_allowance(self: &mut Allowance, asset_addr: address) {
        df::add(&mut self.id, asset_addr.to_bytes(), true);
    }

    // === Remove from the Allowance ===

    /// Removes a balance as a dynamic field of the Allowance.
    public fun remove_coin_allowance<T>(self: &mut Allowance): u64 {
        let coin_type = type_name::get<T>().into_string();
        df::remove<std::ascii::String, u64>(&mut self.id, coin_type)
    }

    /// Removes an asset set as a dynamic field of the Allowance.
    public fun remove_object_allowance(self: &mut Allowance, asset_addr: address): address {
        let exist = df::remove(&mut self.id, asset_addr.to_bytes());
        assert!(exist == true, EAllowanceNotExist);
        asset_addr
    }
}