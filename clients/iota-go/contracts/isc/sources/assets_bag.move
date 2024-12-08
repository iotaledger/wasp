// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
module isc::assets_bag {
    use std::ascii::String;
    use std::type_name;
    use iota::{
        coin::{Self, Coin},
        balance::Balance,
        dynamic_object_field as dof,
        dynamic_field as df,
    };

    // Attempted to destroy a non-empty assets bag
    #[error]
    const EBagNotEmpty: vector<u8> = b"Can't destroy AssetsBag: it still contains assets.";

    // === Main structs ===

    /// An object which allows managing assets within the "ISC" ecosystem.
    /// By default it is owned by a single address.
    public struct AssetsBag has key, store {
        id: UID,
        /// the number of key-value pairs in the bag
        size: u64,
    }

    // === Dynamic Field keys ===

    // === AssetsBag packing and unpacking ===

    /// Creates a new empty `AssetsBag`
    public fun new(ctx: &mut TxContext): AssetsBag {
        AssetsBag{
            id: object::new(ctx),
            size: 0,
        }
    }

    /// Destroys an empty AssetsBag object and returns its balance.
    public fun destroy_empty(self: AssetsBag) {
        let AssetsBag { id, size } = self;
        assert!(size == 0, EBagNotEmpty);
        id.delete();
    }

    // === Place into the AssetsBag ===

    /// Adds a the balance of a Coin as a dynamic field of the AssetsBag where the key is
    /// the type of the Coin (OTW).
    /// Aborts with `EInvalidIOTACoin` if the coin is of type IOTA.
    public fun place_coin<T>(self: &mut AssetsBag, coin: Coin<T>) {
        let balance = coin::into_balance(coin);
        place_coin_balance_internal(self, balance)
    }

    /// Adds a the balance as a dynamic field of the AssetsBag where the key is
    /// the type of the Coin (OTW).
    /// Aborts with `EInvalidIOTACoin` if the coin is of type IOTA.
    public fun place_coin_balance<T>(self: &mut AssetsBag, balance: Balance<T>) {
        place_coin_balance_internal(self, balance)
    }

    /// Adds an asset as a dynamic field of the AssetsBag where the key is
    /// of type Asset (indexed by object id).
    public fun place_asset<T: key + store>(self: &mut AssetsBag, asset: T) {
        place_asset_internal(self, asset)
    }

    // === Take from the AssetsBag ===

    /// Takes an amount from the balance of a Coin set as a dynamic field of the AssetsBag.
    /// Aborts with `EInvalidIOTACoin` if the coin is of type IOTA.
    public fun take_coin_balance<T>(self: &mut AssetsBag, amount: u64): Balance<T> {
        take_coin_balance_internal(self, amount)
    }

    /// Takes all the balance of a Coin set as a dynamic field of the AssetsBag.
    /// Aborts with `EInvalidIOTACoin` if the coin is of type IOTA.
    public fun take_all_coin_balance<T>(self: &mut AssetsBag): Balance<T> {
        let coin_type = type_name::get<T>().into_string();
        self.size = self.size - 1;
        df::remove<String, Balance<T>>(&mut self.id, coin_type)
    }

    /// Takes an asset set as a dynamic field of the AssetsBag.
    public fun take_asset<T: key + store>(self: &mut AssetsBag, id: ID): T {
        take_asset_internal(self, id)
    }

    // === Internal Core ===

    /// Internal: "place" a balance to the AssetsBag.
    /// Aborts with `EInvalidIOTACoin` if the coin is of type IOTA.
    fun place_coin_balance_internal<T>(self: &mut AssetsBag, balance: Balance<T>) {
        let coin_type = type_name::get<T>().into_string();
        if(df::exists_(&self.id, coin_type)) {
            let placed_balance = df::borrow_mut<String, Balance<T>>(&mut self.id, coin_type);
            placed_balance.join(balance);
        } else {
            df::add(&mut self.id, coin_type, balance);
            self.size = self.size + 1;
        }
    }

    /// Internal: "place" an asset to the AssetsBag.
    fun place_asset_internal<T: key + store>(self: &mut AssetsBag, asset: T) {
        dof::add(&mut self.id, object::id(&asset), asset);
        self.size = self.size + 1;
    }

    /// Internal: "take" a balance from the AssetsBag.
    /// Aborts with `EInvalidIOTACoin` if the coin is of type IOTA.
    fun take_coin_balance_internal<T>(self: &mut AssetsBag, amount: u64): Balance<T> {
        let coin_type = type_name::get<T>().into_string();
        let placed_balance = df::borrow_mut<String, Balance<T>>(&mut self.id, coin_type);
        let taken_balance = placed_balance.split(amount);
        if (placed_balance.value() == 0) {
            let zero_balance = df::remove<String, Balance<T>>(&mut self.id, coin_type);
            zero_balance.destroy_zero();
            self.size = self.size - 1;
        };
        taken_balance
    }

    /// Internal: "take" an asset from the AssetsBag.
    fun take_asset_internal<T: key + store>(self: &mut AssetsBag, id: ID): T {
        self.size = self.size - 1;
        dof::remove(&mut self.id, id)
    }
}