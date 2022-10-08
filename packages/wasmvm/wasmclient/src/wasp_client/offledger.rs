// func (c *WaspClient) PostOffLedgerRequest(chainID *isc.ChainID, req isc.OffLedgerRequest) error {
// 	data := model.OffLedgerRequestBody{
// 		Request: model.NewBytes(req.Bytes()),
// 	}
// 	return c.do(http.MethodPost, routes.NewRequest(chainID.String()), data, nil)
// }

use std::borrow::Borrow;

use crate::keypair::*;
use crypto::signatures::ed25519;
use wasmlib::*;

//TODO generalize this trait
pub trait OffLedgerRequest<'a> {
    fn new(
        chain_id: ScChainID,
        contract: ScHname,
        entry_point: ScHname,
        params: ScDict,
        signature_scheme: Option<OffLedgerSignatureScheme>,
        nonce: u64,
    ) -> Self;
    // fn WithNonce(nonce: u64) -> Self;
    // fn WithGasBudget(gasBudget: u64) -> Self;
    fn with_allowance(&self, allowance: &ScAssets) -> &Self;
    fn sign(&self, key: KeyPair) -> &Self;
}

pub struct OffLedgerRequestData {
    chain_id: ScChainID,
    contract: ScHname,
    entry_point: ScHname,
    params: ScDict,
    signature_scheme: Option<OffLedgerSignatureScheme>, // None if unsigned
    nonce: u64,
    allowance: ScAssets,
    gas_budget: u64,
}

pub struct OffLedgerSignatureScheme {
    public_key: ed25519::PublicKey,
    signature: Vec<u8>,
}

impl OffLedgerRequest<'_> for OffLedgerRequestData {
    fn new(
        chain_id: ScChainID,
        contract: ScHname,
        entry_point: ScHname,
        params: ScDict,
        signature_scheme: Option<OffLedgerSignatureScheme>,
        nonce: u64,
    ) -> Self {
        return OffLedgerRequestData {
            chain_id: chain_id,
            contract: contract,
            entry_point: entry_point,
            params: params,
            signature_scheme: signature_scheme,
            nonce: nonce,
            allowance: ScAssets::new(&Vec::new()),
            gas_budget: super::gas::MAX_GAS_PER_REQUEST,
        };
    }
    fn with_allowance(&self, allowance: &ScAssets) -> &Self {
        return self;
    }
    fn sign(&self, key: KeyPair) -> &Self {
        return self;
    }
}
