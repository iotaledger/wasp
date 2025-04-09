// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::anchor_tests {

    use std::{
        string::{Self},
        ascii::Self,
    };
    use iota::{
        coin::Self,
        iota::IOTA,
        url::Self,
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

    use std::fixed_point32;
    use iota::vec_map;


    // One Time Witness for coins used in the tests.
    public struct TEST_A has drop {}

    // Demonstration on how to receive a Request inside one PTB.
    #[test]
    fun demonstrate_request_ptb() {
        // Setup
        let initial_iota_in_request = 10000;
        let initial_testA_in_request = 100;
        let chain_owner = @0xA;
        let sender = @0xB;
        let mut ctx = tx_context::dummy();

        // Create an Anchor.
        let mut anchor = anchor::start_new_chain(vector::empty(), option::none(), &mut ctx);

        // ClientPTB.1 Mint some tokens for the request.
        let iota = coin::mint_for_testing<IOTA>(initial_iota_in_request, &mut ctx);
        let test_a_coin = coin::mint_for_testing<TEST_A>(initial_testA_in_request, &mut ctx);

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
            &mut ctx,
        );

        let nft_id = object::id(&test_b_nft);

        // ClientPTB.2 create the allowance (empty means no allowance)
        let allowance = vector::empty();

        // ClientPTB.3 Add the assets to the bag.
        let mut req_assets = assets_bag::new(&mut ctx);
        req_assets.place_coin(test_a_coin);
        req_assets.place_coin(iota);
        req_assets.place_asset(test_b_nft);

        // ClientPTB.4 Create the request and can send it to the Anchor.
        /*request::create_and_send_request(
            object::id(&anchor).id_to_address(),
            req_assets,
            option::some(string::utf8(b"contract")), 
            option::some(string::utf8(b"function")), 
            option::some(vector::empty()), 
            &mut ctx,
        );*/ // Commented because cannot be executed received in this test
        // Instead create a test request
        let req = request::create_for_testing(
            req_assets,
            42, // contract hname
            42, // entry point
            vector::empty(), // args
            allowance,
            100,
            &mut ctx,
        );

        // ServerPTB.1 Now the Anchor receives off-chain an event that tracks the request and can receive it.
        // let (receipt, req_extracted_assets) = anchor.receive_request(req); // Commented because cannot be executed in this test
        let (id, mut req_extracted_assets) = req.destroy(); //this is not part of the PTB
        let receipt = anchor::create_receipt_for_testing(id); //this is not part of the PTB

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
        assert!(extracted_iota_balance.value() == 10000);
        // ServerPTB.3.2: place it to the anchor assets bag.
        anchor_assets.place_coin_balance(extracted_iota_balance);

        // ServerPTB.4.1: extract the nft.
        let extracted_nft = req_extracted_assets.take_asset<Nft>(nft_id);
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

        transfer::public_transfer(anchor, chain_owner); // not needed in the PTB
    }

    #[test]
    fun test_anchor_state_update() {
        let chain_owner = @0xA;
        let mut ctx = tx_context::dummy();

        let mut anchor = anchor::start_new_chain(vector::empty(), option::none(), &mut ctx);

        let metadata = vector[1,2,3];
        anchor.update_anchor_state_for_migration(metadata, 1);

        assert!(anchor.get_state_index() == 1);
        assert!(anchor.get_state_metadata() == metadata);

        transfer::public_transfer(anchor, chain_owner); 
    }

    #[test]
    fun test_migration_asset_placement() {
        let initial_iota_in_request = 10000;
        let initial_testA_in_request = 100;
        let chain_owner = @0xA;
        let mut ctx = tx_context::dummy();

        let mut anchor = anchor::start_new_chain(vector::empty(), option::none(), &mut ctx);

        // Mint some tokens for the request.
        let iota = coin::mint_for_testing<IOTA>(initial_iota_in_request, &mut ctx);
        let test_a_coin = coin::mint_for_testing<TEST_A>(initial_testA_in_request, &mut ctx);

        // Place tokens into the anchors AssetsBag using the migration place functions.
        anchor.place_coin_for_migration(iota);
        anchor.place_coin_for_migration(test_a_coin);

        // Take back the assets out of the AssetsBag
        let (mut assets,b) = anchor.borrow_assets();
        let iota_coin = assets.take_coin_balance<IOTA>(initial_iota_in_request);
        let test_coin = assets.take_coin_balance<TEST_A>(initial_testA_in_request);

        // Make sure that the balance remained the same
        assert!(iota_coin.value() == initial_iota_in_request);
        assert!(test_coin.value() == initial_testA_in_request);
        assert!(assets.get_size() == 0);
        
        anchor.return_assets_from_borrow(assets, b);

        // Clean up
        transfer::public_transfer(anchor, chain_owner); 
        
        let mut temp_assets_bag = assets_bag::new(&mut ctx);
        temp_assets_bag.place_coin_balance(test_coin);
        temp_assets_bag.place_coin_balance(iota_coin);
        transfer::public_transfer(temp_assets_bag, chain_owner); // not needed in the PTB
    }
}
