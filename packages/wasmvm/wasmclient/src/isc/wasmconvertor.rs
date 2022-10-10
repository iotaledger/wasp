pub use crate::types::*;
pub struct WasmConvertor {}

impl WasmConvertor {
    pub fn new() -> Self {
        todo!()
    }

    pub fn isc_address(&self, addr: &wasmlib::ScAddress) -> Address {
        todo!()
        // return wasmlib.bytesToUint8Array(wasmlib.addressToBytes(addr));
    }

    pub fn isc_allowance(&self, addr: &wasmlib::ScAssets) -> Allowance {
        todo!()
    }

    pub fn isc_chain_id(&self, chain_id: &wasmlib::ScChainID) -> ChainID {
        todo!()
        // return wasmlib.bytesToUint8Array(wasmlib.chainIDToBytes(chainID));
    }

    pub fn isc_hname(&self, h_name: &wasmlib::ScHname) -> Hname {
        todo!()
        // return wasmlib.uint32FromBytes(wasmlib.hnameToBytes(hName));
    }

    pub fn isc_request_id(&self, chain_id: &wasmlib::ScRequestID) -> RequestID {
        todo!()
        // return wasmlib.bytesToUint8Array(wasmlib.requestIDToBytes(chainID));
    }

    pub fn sc_address(&self, addr: Address) -> wasmlib::ScAddress {
        todo!()
        // return wasmlib.addressFromBytes(wasmlib.bytesFromUint8Array(addr));
    }

    pub fn sc_chain_id(&self, chain_id: ChainID) -> wasmlib::ScChainID {
        todo!()
        // return wasmlib.chainIDFromBytes(wasmlib.bytesFromUint8Array(chainID));
    }

    pub fn sc_hname(&self, h_name: Hname) -> wasmlib::ScHname {
        todo!()
        // return wasmlib.hnameFromBytes(wasmlib.uint32ToBytes(hName));
    }

    pub fn sc_request_id(&self, request_id: RequestID) -> wasmlib::ScRequestID {
        todo!()
        // return wasmlib.requestIDFromBytes(wasmlib.bytesFromUint8Array(requestID));
    }
}
