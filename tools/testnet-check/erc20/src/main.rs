use alloy::network::{ReceiptResponse, TransactionBuilder};
use alloy::providers::{Provider, ProviderBuilder};
use alloy::{rpc::types::TransactionRequest, signers::local::PrivateKeySigner, sol};
use alloy_primitives::U256;
use eyre::Result;
use std::str::FromStr;

// Codegen from artifact.
sol!(
    #[allow(missing_docs)]
    #[sol(rpc)]
    TestToken,
    "out/TestToken.sol/TestTokenERC20.json"
);

#[tokio::main]
async fn main() -> Result<()> {
    let rpc_url = "https://api.evm.lb-0.h.testnet.iota.cafe/v1/chain/evm";
    let alice = PrivateKeySigner::from_str(
        "a funded private key",
    )
    .unwrap();
    let alice_address = &alice.address();
    println!("alice_address: {alice_address}");
    let bob = PrivateKeySigner::random();
    let bob_address = &bob.address();
    println!("bob_address: {bob_address}");
    let provider = ProviderBuilder::new()
        .wallet(alice.clone())
        .on_http(reqwest::Url::parse(rpc_url).unwrap());

    let bytecode_str = std::fs::read_to_string("out/TestToken.sol/TestTokenERC20.json")?;
    let json: serde_json::Value = serde_json::from_str(&bytecode_str)?;
    let bytecode_hex = json["bytecode"]["object"].as_str().unwrap().to_string();
    let bytecode = hex::decode(&bytecode_hex[2..bytecode_hex.len()]).unwrap();

    let tx = TransactionRequest::default()
        .from(alice_address.clone())
        .with_deploy_code(bytecode)
        .with_gas_price(1_000_000_000u128);

    let gas = provider.estimate_gas(tx.clone()).await.unwrap();
    println!("gas: {gas}");

    // Deploy the contract.
    let receipt = provider.send_transaction(tx).await?.get_receipt().await?;

    let contract_address = receipt.contract_address.unwrap();
    let contract = TestToken::new(contract_address, &provider);

    let contract_name = contract.name().call().await?;
    let contract_symbol = contract.symbol().call().await?;
    let contract_total_supply = contract.totalSupply().call().await?;
    assert_eq!("TestToken", contract_name);
    assert_eq!("TTK", contract_symbol);
    assert_eq!("1000000", contract_total_supply.to_string());

    // Register the balances of Alice and Bob before the transfer.
    let alice_before_balance = contract.balanceOf(alice_address.clone()).call().await?;
    let bob_before_balance = contract.balanceOf(bob_address.clone()).call().await?;
    let gas_price = provider.get_gas_price().await?;

    // Transfer and wait for inclusion.
    let amount = U256::from(100);
    let tx_hash = contract
        .transfer(bob_address.clone(), amount)
        .gas(gas)
        .gas_price(gas_price)
        .send()
        .await?
        .watch()
        .await?;

    println!("Sent transaction: {tx_hash}");

    // Register the balances of Alice and Bob after the transfer.
    let alice_after_balance = contract.balanceOf(alice_address.clone()).call().await?;
    let bob_after_balance = contract.balanceOf(bob_address.clone()).call().await?;

    // Check the balances of Alice and Bob after the transfer.
    assert_eq!(alice_before_balance - alice_after_balance, amount);
    assert_eq!(bob_after_balance - bob_before_balance, amount);
    println!("all good");
    Ok(())
}
