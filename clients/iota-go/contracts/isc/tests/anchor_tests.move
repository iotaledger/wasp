// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::anchor_tests {

    use std::{
        fixed_point32,
        string::{Self, String},
        ascii::Self,
    };
    use iota::{
        table::Self,
        coin::Self,
        coin::Coin,
        iota::IOTA,
        url::Self,
        vec_set::Self,
        vec_map,
        test_utils,
        test_scenario,
    };

    use stardust::{
        nft::{Self, Nft},
        irc27::Self,
    };

    use isc::{
        assets_bag::Self,
        anchor::Self,
        request::Self,
    };

    // One Time Witness for coins used in the tests.
    public struct TEST_A has drop {}
    public struct TEST_B has drop {}

    // Demonstration on how to receive a Request inside one PTB.

    public fun create_fake_nft(sender: address, ctx: &mut TxContext): stardust::nft::Nft {
        let mut royalties = vec_map::empty();
        royalties.insert(sender, fixed_point32::create_from_rational(1, 2));

        let mut attributes = vec_map::empty();
        attributes.insert(string::utf8(b"attribute"), string::utf8(b"value"));

        let mut non_standard_fields = vec_map::empty();
        non_standard_fields.insert(string::utf8(b"field"), string::utf8(b"value"));

        let test_b_nft = nft::create_for_testing(
            option::some(sender),
            option::some(b"metadata"),
            option::some(b"tag"),
            option::some(sender),
            irc27::create_for_testing(
                string::utf8(b"0.0.1"),
                string::utf8(b"image/png"),
                url::new_unsafe(ascii::string(b"www.best-nft.com/nft.png")),
                string::utf8(b"nft"),
                option::some(string::utf8(b"collection")),
                royalties,
                option::some(string::utf8(b"issuer")),
                option::some(string::utf8(b"description")),
                attributes,
                non_standard_fields,
            ),
             ctx,
        );

        test_b_nft
    }


    #[test]
    fun demonstrate_request_ptb() {
        // Setup
        
        let initial_iota_in_request = 10000;
        let initial_testA_in_request = 100;
        let governor = @0xA;
        let sender = @0xB;
        let mut ctx = tx_context::dummy();

        let coin: Option<Coin<IOTA>> = option::none();

        let mut iota_gas_coin = coin::mint_for_testing<IOTA>(0, &mut ctx);
        let gas_coin_id = object::borrow_id(&iota_gas_coin);
        let gas_coin_addr = object::id_to_address(gas_coin_id);

        // Create an Anchor.
        let mut anchor = anchor::start_new_chain(vector::empty(), gas_coin_addr, coin, 16, governor, &mut ctx);

        // ClientPTB.1 Mint some tokens for the request.
        let mut iota = coin::mint_for_testing<IOTA>(initial_iota_in_request, &mut ctx);
        let test_a_coin = coin::mint_for_testing<TEST_A>(initial_testA_in_request, &mut ctx);

        let test_b_nft_obj = create_fake_nft(sender, &mut ctx);
        let test_b_nft_id = object::id(&test_b_nft_obj);

        // ClientPTB.2 create allowance
        let mut allowance_cointypes = vector::empty();
        let mut allowance_balances = vector::empty();
        allowance_cointypes.push_back(string::utf8(b"TEST_A"));
        allowance_balances.push_back(100);
        allowance_cointypes.push_back(string::utf8(b"TEST_A"));
        allowance_balances.push_back(111);
        allowance_cointypes.push_back(string::utf8(b"IOTA"));
        allowance_balances.push_back(32);

        // ClientPTB.3 Add the assets to the bag.
        let mut req_assets = assets_bag::new(&mut ctx);
        req_assets.place_coin(test_a_coin);
        req_assets.place_coin(iota);
        req_assets.place_asset(test_b_nft_obj);

        let mut scenario = test_scenario::begin(sender);
        // ClientPTB.4 Create the request and can send it to the Anchor.
        let request_id = request::create_and_send_request(
            object::id(&anchor).id_to_address(),
            req_assets,
            42, 
            42, 
            vector::empty(), // args
            allowance_cointypes,
            allowance_balances,
            100,
            &mut ctx,
        );

        test_scenario::next_tx(&mut scenario, sender);
        // ServerPTB.1 Now the Anchor receives off-chain an event that tracks the request and can receive it.
        std::debug::print(&request_id);
        let req = test_scenario::most_recent_receiving_ticket<isc::request::Request>(&object::id(&anchor));
        std::debug::print(&req);
        scenario.end();

        assert!(request_id == req.receiving_object_id());

        let (receipt, mut req_extracted_assets) = anchor.receive_request(req); 
        std::debug::print(&receipt);
        std::debug::print(&req_extracted_assets);
        

        // ServerPTB.2: borrow the asset bag of the anchor
        let (mut anchor_assets, borrow) = anchor.borrow_assets();    

        // ServerPTB.2.1: extract half the balance A.
        let extracted_test_a_coin_balance_half = req_extracted_assets.take_coin_balance<TEST_A>(initial_testA_in_request / 2);

        // ServerPTB.2.2: place it to the anchor assets bag.
        anchor_assets.place_coin_balance(extracted_test_a_coin_balance_half);
        // ServerPTB.2.3: extract all the balance A.
        let extracted_test_a_coin_balance = req_extracted_assets.take_all_coin_balance<TEST_A>();
        // ServerPTB.2.4: place it to the anchor assets bag.
        anchor_assets.place_coin_balance(extracted_test_a_coin_balance);
        
        // ServerPTB.3.1: extract the iota balance.
        let extracted_iota_balance = req_extracted_assets.take_all_coin_balance<IOTA>();
        std::debug::print(&extracted_iota_balance.value());

        assert!(extracted_iota_balance.value() == 10000);
        // ServerPTB.3.2: place it to the anchor assets bag.
        anchor_assets.place_coin_balance(extracted_iota_balance);

        // ServerPTB.4.1: extract the nft.
        let extracted_nft = req_extracted_assets.take_asset<Nft>(test_b_nft_id);
        // ServerPTB.4.2: place it to the anchor assets bag.
        anchor_assets.place_asset(extracted_nft);

        // ServerPTB.5: destroy the request assets bag.
        req_extracted_assets.destroy_empty();

        // ServerPTB.6: return the anchor assets bag from the borrow.
        anchor.return_assets_from_borrow(anchor_assets, borrow);

        // ServerPTB.7: update the state root and destroy the hot potato receipt.
        // ServerPTB.7.1: create the receipts vector.
        let mut receipts = vector::empty();
        receipts.push_back(receipt);

        // ServerPTB.7.2: update the state root
        let new_state_metadata = vector::empty();
        anchor.transition(new_state_metadata, receipts);

        // !!! END !!!

        transfer::public_transfer(anchor, governor); // not needed in the PTB

        iota_gas_coin.destroy_zero();
    }
}
