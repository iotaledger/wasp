pub struct Receipt {
    request: &'static [u8],
    error: Result<(), &'static str>,
    gas_budget: u64,
    gas_burned: u64,
    gas_fee_charged: u64,
    block_index: u32,
    request_index: u16,
    resolved_error: &'static str,
    gas_burn_log: *const BurnLog,
}

type BurnCode = u16;

pub struct BurnRecord {
    code: BurnCode,
    gas_burned: u64,
}

pub struct BurnLog {
    records: [BurnRecord],
}
