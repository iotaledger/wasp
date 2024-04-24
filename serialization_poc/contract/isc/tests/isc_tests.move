
#[test_only]
module isc::isc_tests {
    // uncomment this line to import the module
    use isc::anchor;
    use sui::test_scenario;

    use std::debug;
    const ENotImplemented: u64 = 0;

    #[test]
    fun test_isc() {
        let initial_owner = @0xCAFE;
        let dummy_address = @0xC0FFEE;
        let mut scenario = test_scenario::begin(initial_owner);
        
        let (res, a, b) = isc::anchor::start_new_chain(scenario.ctx());
        
        debug::print(&res);
        debug::print(&a);
        debug::print(&b);

        transfer::public_transfer(res, dummy_address);
        transfer::public_transfer(a, dummy_address);
        transfer::public_transfer(b, dummy_address);

        scenario.next_tx(initial_owner);
        scenario.end();
    }

    #[test, expected_failure(abort_code = isc::isc_tests::ENotImplemented)]
    fun test_isc_fail() {
        abort ENotImplemented
    }
}
