use alloy::network::{ReceiptResponse, TransactionBuilder};
use alloy::providers::{Provider, ProviderBuilder};
use alloy::{rpc::types::TransactionRequest, signers::local::PrivateKeySigner, sol};
use alloy_primitives::{Address, U256};
use eyre::Result;
use std::str::FromStr;

// Codegen from artifact.
sol!(
    #[allow(missing_docs)]
    #[sol(rpc)]
    TestNFT,
    "out/TestNFT.sol/TestNFT.json"
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

    let mut call_builder = TestNFT::deploy_builder(provider.clone(), alice_address.clone());
    let tx = call_builder.clone().into_transaction_request();
    let gas = provider.estimate_gas(tx.clone()).await.unwrap();
    println!("gas: {gas}");
    let gas_price = provider.get_gas_price().await?;
    call_builder = call_builder.gas(gas).gas_price(gas_price);
    let receipt = call_builder.send().await?.get_receipt().await?;

    let contract_address = receipt.contract_address.unwrap();
    let contract = TestNFT::new(contract_address, &provider);

    let contract_name = contract.name().call().await?;
    let contract_symbol = contract.symbol().call().await?;
    println!("contract_name: {contract_name}");
    println!("contract_symbol: {contract_symbol}");

    // Mint a token to Alice
    let tx_hash = contract
        .safeMint(alice_address.clone())
        .gas(gas)
        .gas_price(gas_price)
        .send()
        .await?
        .watch()
        .await?;
    println!("Minted token to Alice: {:?}", tx_hash);

    // Check ownerOf(0)
    let owner = contract.ownerOf(U256::from(0)).call().await?;
    println!("Owner of token 0: {owner}");
    assert_eq!(owner, alice_address.clone());

    // Transfer token 0 from Alice to Bob
    let transfer_tx = contract
        .transferFrom(alice_address.clone(), bob_address.clone(), U256::from(0))
        .gas(gas)
        .gas_price(gas_price)
        .send()
        .await?
        .watch()
        .await?;
    println!("Transferred token 0 from Alice to Bob: {:?}", transfer_tx);

    // Verify new owner
    let new_owner = contract.ownerOf(U256::from(0)).call().await?;
    println!("New owner of token 0: {new_owner}");
    assert_eq!(new_owner, bob_address.clone());

    println!("all good");
    Ok(())
}
