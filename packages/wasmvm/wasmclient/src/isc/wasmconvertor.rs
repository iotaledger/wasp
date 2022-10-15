pub use crate::types::*;
pub use wasmlib::wasmtypes;

pub struct WasmConvertor {}

impl WasmConvertor {
    pub fn new() -> Self {
        return WasmConvertor {};
    }

    pub fn isc_address(&self, addr: &wasmlib::ScAddress) -> Address {
        return wasmlib::address_to_bytes(&addr) as Address;
    }

    pub fn isc_allowance(&self, addr: &wasmlib::ScAssets) -> Allowance {
        // FIXME
        return addr.to_bytes() as Allowance;
    }

    pub fn isc_chain_id(&self, chain_id: &wasmlib::ScChainID) -> ChainID {
        return wasmlib::chain_id_to_bytes(&chain_id) as ChainID;
    }

    pub fn isc_hname(&self, hname: &wasmlib::ScHname) -> Hname {
        return wasmtypes::uint32_from_bytes(&wasmlib::hname_to_bytes(&hname)) as Hname;
    }

    pub fn isc_request_id(&self, request_id: &wasmlib::ScRequestID) -> RequestID {
        return wasmlib::request_id_to_bytes(&request_id) as RequestID;
    }

    pub fn sc_address(&self, addr: &Address) -> wasmlib::ScAddress {
        return wasmlib::address_from_bytes(addr);
    }

    pub fn sc_chain_id(&self, chain_id: &ChainID) -> wasmlib::ScChainID {
        return wasmlib::chain_id_from_bytes(chain_id);
    }

    pub fn sc_hname(&self, hname: &Hname) -> wasmlib::ScHname {
        return wasmlib::hname_from_bytes(&wasmtypes::uint32_to_bytes(hname.clone()));
    }

    pub fn sc_request_id(&self, request_id: &RequestID) -> wasmlib::ScRequestID {
        return wasmlib::request_id_from_bytes(request_id);
    }
}
