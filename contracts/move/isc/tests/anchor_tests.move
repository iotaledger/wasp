// Copyright (c) 2024 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

module isc::anchor_tests {

    use std::{
        fixed_point32::{FixedPoint32},
        string::{Self, String},
        ascii::{Self},
    };
    use sui::{
        table::{Self},
        coin::{Self},
        sui::SUI,
        url::{Self},
        vec_set::{Self},
    };

    use stardust::{
        nft::{Self, Nft},
        irc27::{Self},
    };
    use isc::{
        assets_bag::{Self},
        anchor::{Self},
        request::{Self},
    };

    // One Time Witness for coins used in the tests.
    public struct TEST_A has drop {}
    public struct TEST_B has drop {}

    // Demonstration on how to receive a Request inside one PTB.
    #[test]
    fun demonstrate_request_ptb() {
        // Setup
        let initial_iota_in_request = 10000;
        let initial_testA_in_request = 100;
        let governor = @0xA;
        let sender = @0xB;
        let mut ctx = tx_context::dummy();

        // Create an Anchor.
        let mut anchor = anchor::start_new_chain(&mut ctx);

        // ClientPTB.1 Mint some tokens for the request.
        let iota = coin::mint_for_testing<SUI>(initial_iota_in_request, &mut ctx);
        let test_a_coin = coin::mint_for_testing<TEST_A>(initial_testA_in_request, &mut ctx);
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
                table::new<address,FixedPoint32>(&mut ctx),
                option::some(string::utf8(b"issuer")),
                option::some(string::utf8(b"description")),
                vec_set::empty<String>(),
                table::new<String,String>(&mut ctx),
            ),
            &mut ctx,
        );
        let nft_id = object::id(&test_b_nft);

        // ClientPTB.2 Add the assets to the bag.
        let mut req_assets = assets_bag::new(&mut ctx);
        req_assets.place_coin(test_a_coin);
        req_assets.place_coin(iota);
        req_assets.place_asset(test_b_nft);

        // ClientPTB.3. Create the request and can send it to the Anchor.
        /*request::create_and_send_request(
            object::id(&anchor).id_to_address(),
            req_assets,
            option::some(string::utf8(b"contract")), 
            option::some(string::utf8(b"function")), 
            option::some(vector::empty()), 
            &mut ctx,
        );*/ // Commented because cannot be executed received in this test
        let req = request::create_for_testing(
            req_assets,
            option::some(string::utf8(b"contract")), 
            option::some(string::utf8(b"function")), 
            option::some(vector::empty()), 
            &mut ctx,
        );

        // ServerPTB.1 Now the Anchor receives off-chain an event that tracks the request and can receive it.
        //let (receipt, req_extracted_assets) = anchor.receive_request(req); // Commented because cannot be executed in this test
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
        let extracted_iota_balance = req_extracted_assets.take_all_coin_balance<SUI>();
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
        anchor.update_state_root(vector::empty(), receipts);

        // !!! END !!!

        transfer::public_transfer(anchor, governor); // not needed in the PTB
    }
}