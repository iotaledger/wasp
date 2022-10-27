// func (c *WaspClient) PostOffLedgerRequest(chainID *isc.ChainID, req isc.OffLedgerRequest) error {
// 	data := model.OffLedgerRequestBody{
// 		Request: model.NewBytes(req.Bytes()),
// 	}
// 	return c.do(http.MethodPost, routes.NewRequest(chainID.String()), data, nil)
// }

use crate::keypair::*;
use crypto::signatures::ed25519;
use wasmlib::*;

//TODO generalize this trait
pub trait OffLedgerRequest {
    fn new(
        chain_id: &ScChainID,
        contract: &ScHname,
        entry_point: &ScHname,
        params: &ScDict,
        signature_scheme: Option<&OffLedgerSignatureScheme>,
        nonce: u64,
    ) -> Self;
    fn with_nonce(&mut self, nonce: u64) -> &Self;
    fn with_gas_budget(&mut self, gas_budget: u64) -> &Self;
    fn with_allowance(&mut self, allowance: &ScAssets) -> &Self;
    fn sign(&mut self, key: &KeyPair) -> &Self;
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

#[derive(Clone)]
pub struct OffLedgerSignatureScheme {
    public_key: ed25519::PublicKey,
    signature: Vec<u8>,
}

impl OffLedgerRequest for OffLedgerRequestData {
    fn new(
        chain_id: &ScChainID,
        contract: &ScHname,
        entry_point: &ScHname,
        params: &ScDict,
        signature_scheme: Option<&OffLedgerSignatureScheme>,
        nonce: u64,
    ) -> Self {
        return OffLedgerRequestData {
            chain_id: chain_id.clone(),
            contract: contract.clone(),
            entry_point: entry_point.clone(),
            params: params.clone(),
            signature_scheme: match signature_scheme {
                Some(val) => Some(val.clone()),
                None => None,
            },
            nonce: nonce,
            allowance: ScAssets::new(&Vec::new()),
            gas_budget: super::gas::MAX_GAS_PER_REQUEST,
        };
    }
    fn with_nonce(&mut self, nonce: u64) -> &Self {
        self.nonce = nonce;
        return self;
    }
    fn with_gas_budget(&mut self, gas_budget: u64) -> &Self {
        self.gas_budget = gas_budget;
        return self;
    }
    fn with_allowance(&mut self, allowance: &ScAssets) -> &Self {
        self.allowance = allowance.clone();
        return self;
    }
    fn sign(&mut self, _key: &KeyPair) -> &Self {
        todo!()
    }
}

impl OffLedgerRequestData {
    pub fn id(&self) -> ScRequestID {
        todo!()
    }
}
