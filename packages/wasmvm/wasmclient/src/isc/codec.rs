pub use crate::types::*;

pub const bech32_prefix: &'static str = "smr";

pub fn bech32_decode(bech32: &str) -> Result<Address, String> {
    todo!()
    // let dec = Bech32.decode(bech32);
    // if (dec == undefined) {
    //     return null;
    // }
    // return dec.data;
}

pub fn bech32_encode(addr: &Address) -> String {
    todo!()
    // return Bech32.encode(Codec.bech32Prefix, addr);
}

pub fn hname_bytes(name: &str) -> Vec<u8> {
    todo!()
    // const data = Uint8Array.wrap(String.UTF8.encode(name));
    // let hash = Blake2b.sum256(data)

    // // follow exact algorithm from packages/isc/hname.go
    // let slice = wasmlib.bytesFromUint8Array(hash.slice(0, 4));
    // let hName = wasmlib.uint32FromBytes(slice);
    // if (hName == 0 || hName == 0xffff) {
    //     slice = wasmlib.bytesFromUint8Array(hash.slice(4, 8));
    // }
    // return slice;
}
