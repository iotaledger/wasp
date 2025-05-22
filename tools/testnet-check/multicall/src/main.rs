//! This example demonstrates how to use the [`MulticallBuilder`] to make multicalls using the
//! [`IMulticall3`] contract.

use alloy::network::{ReceiptResponse, TransactionBuilder};
use alloy::{
    primitives::{address, U256},
    providers::{CallItemBuilder, Failure, Provider, ProviderBuilder},
    signers::local::PrivateKeySigner,
    sol,
};
use std::str::FromStr;

sol!(
    #[allow(missing_docs)]
    #[sol(rpc)]
    #[derive(Debug)]
    IWETH9,
    "out/WETH9.sol/WETH9.json"
);

sol!(
    #[allow(missing_docs)]
    #[sol(rpc)]
    #[derive(Debug)]
    Multicall3,
    "out/Multicall3.sol/Multicall3.json"
);

#[tokio::main]
async fn main() -> eyre::Result<()> {
    // Create a new provider
    let rpc_url = "https://api.evm.lb-0.h.testnet.iota.cafe/v1/chain/evm";
    let alice = PrivateKeySigner::from_str(
        "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
    )
    .unwrap();
    let provider = ProviderBuilder::new()
        .wallet(alice.clone())
        .on_http(reqwest::Url::parse(rpc_url).unwrap());
    // Create a new instance of the IWETH9 contract.
    let weth = IWETH9::new(
        address!("0x349F71aE4cC64D5c6Def4eAf964467AAFDc3Da10"),
        &provider,
    );

    let bob = PrivateKeySigner::random();
    let alice_address = &alice.address();
    let bob_address = &bob.address();

    let mut call_builder = Multicall3::deploy_builder(provider.clone());
    let tx = call_builder.clone().into_transaction_request();
    let gas = provider.estimate_gas(tx.clone()).await.unwrap();
    println!("gas: {gas}");
    let gas_price = provider.get_gas_price().await?;
    call_builder = call_builder.gas(gas).gas_price(gas_price);
    let receipt = call_builder.send().await?.get_receipt().await?;

    let contract_address = receipt.contract_address.unwrap();
    println!("contract_address: {contract_address}");

    let multicall = provider
        .multicall()
        .address(contract_address.clone())
        // Set the address of the Multicall3 contract. If unset it uses the default address from <https://github.com/mds1/multicall>: 0xcA11bde05977b3631167028862bE2a173976CA11
        // .address(multicall3)
        // Get the total supply of WETH on our anvil fork.
        .add(weth.totalSupply())
        // Get Alice's WETH balance.
        .add(weth.balanceOf(alice_address.clone()))
        // Also fetch Alice's ETH balance.
        .get_eth_balance(alice_address.clone());

    let (init_total_supply, alice_weth, alice_eth_bal) = multicall.aggregate().await?;

    println!(
        "Initial total supply: {}, Alice's WETH balance: {}, Alice's ETH balance: {}",
        init_total_supply, alice_weth, alice_eth_bal
    );

    // Simulate a transfer of WETH from Alice to Bob.
    let wad = U256::from(20);

    // This would fail as Alice doesn't have any WETH.
    let tx = CallItemBuilder::new(weth.transfer(bob_address.clone(), U256::from(10)))
        .allow_failure(true);
    let deposit = CallItemBuilder::new(weth.deposit()).value(wad); // Set the amount of eth that should be deposited into the contract.
    let multicall = provider
        .multicall()
        .address(contract_address.clone())
        // Bob's intial WETH balance.
        .add(weth.balanceOf(bob_address.clone()))
        // Attempted WETH transfer from Alice to Bob which would fail.
        .add_call(tx.clone())
        // Alices deposits ETH and mints WETH.
        .add_call(deposit)
        // Attempt transfer again. Succeeds!
        .add_call(tx)
        // Alice's WETH balance after the transfer.
        .add(weth.balanceOf(alice_address.clone()))
        // Bob's final balance.
        .add(weth.balanceOf(bob_address.clone()));

    assert_eq!(multicall.len(), 6);

    // It is important to use `aggregate3_value` as we're trying to simulate calls to payable
    // functions that should be sent a value, using any other multicall3 method would result in an
    // error.
    let (init_bob, failed_transfer, deposit, succ_transfer, alice_weth, bob_weth) =
        multicall.aggregate3_value().await?;

    // Since, `aggregate3_value` allows for calls to fail without reverting, it returns a tuple of
    // results which contain the decoded return value in Ok(_) variant and the `Failure` type in the
    // Err(_) variant.
    assert!(matches!(
        failed_transfer.unwrap_err(),
        Failure {
            idx: 1,
            return_data: _
        }
    ));

    let init_bob = init_bob?;
    assert_eq!(init_bob, U256::ZERO);

    assert!(deposit.is_ok());
    assert!(succ_transfer.is_ok());

    let alice_weth = alice_weth?;
    let bob_weth = bob_weth?;

    println!(
        "Alice's WETH balance: {}, Bob's WETH balance: {}",
        alice_weth, bob_weth
    );

    Ok(())
}
