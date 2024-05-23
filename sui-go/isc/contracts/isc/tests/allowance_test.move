#[test_only]
module isc::allowance_test {
    use isc::allowance::{Self,Allowance};
    use std::type_name;
    use sui::sui::SUI;

    #[test]
    fun test_allowance() {
        let mut ctx = tx_context::dummy();

        let uid = object::new(&mut ctx);
        let nft_id = uid.to_inner();
        object::delete(uid);
        
        let uid2 = object::new(&mut ctx);
        let nft_id2 = uid2.to_inner();
        object::delete(uid2);
        
        let mut a = allowance::new();
    
        let name = type_name::get<SUI>().into_string();
        std::debug::print(&name);
        a.add_coin(&name, 100);

        a.add_nft(nft_id);

        assert!(a.get_coin_amount(&name) == 100, 2);
        let dummy = type_name::get<Allowance>().into_string();
        assert!(a.get_coin_amount(&dummy) == 0, 3);
        assert!(a.has_nft(nft_id), 4);
        assert!(!a.has_nft(nft_id2), 5);
    }

    #[test, expected_failure(abort_code = isc::allowance::EDuplicateNft)]
    fun test_allowance_duplicate_nft() {
        let mut ctx = tx_context::dummy();

        let uid = object::new(&mut ctx);
        let nft_id = uid.to_inner();
        object::delete(uid);

        let mut a = allowance::new();
        a.add_nft(nft_id);
        a.add_nft(nft_id);
    }
}
