import { Ed25519Keypair } from "@iota/iota-sdk/keypairs/ed25519";
import { bcs } from "@iota/iota-sdk/bcs";
import { IOTA_TYPE_ARG, toHex } from "@iota/iota-sdk/utils";
import { getHname, IscAgentID, IscTransaction } from "@iota/isc-sdk";
import { ChainsApi, createConfiguration, ServerConfiguration } from "./client/";

const PACKAGE_ID =
  "0xab31cace7321b60db8322c667590645333eedc812a81e17a64550c7a0b3e2476";
const CHAIN_ID =
  "0xcbb034f9c59a4a3cd2131d30229c611de55d641a3400872f442a539cf7af74c7";

function agentIdForEVM(address: string): Uint8Array {
  const agentID = IscAgentID.serialize({
    EthereumAddressAgentID: {
      eth: bcs.fixedArray(20, bcs.u8()).fromHex(address),
    },
  });

  return agentID.toBytes();
}

async function estimateOnLedger(txBytes: Uint8Array<ArrayBufferLike>) {
  const client = new ChainsApi(
    createConfiguration({
      baseServer: new ServerConfiguration("http://localhost:9090", {}), // ISC API URL
    })
  );

  const estimation = await client.estimateGasOnledger({
    transactionBytes: "0x" + toHex(txBytes),
  });

  console.log(estimation);
}

// This function creates a "transferAllowanceTo" transaction
// This will deposit L1 funds into a L2 EVM target address.
// The intention here is to estimate a transaction without owning any assets on neither L1 or L2. 
// Therefore, the tx gets built with preset Gas data to enable clientless tx building.
async function createTestTransaction(
  wallet: Ed25519Keypair
) {
  const iscTx = new IscTransaction({
    chainId: CHAIN_ID,
    packageId: PACKAGE_ID,
  });

  const bag = iscTx.newBag();
  const coin = iscTx.coinFromAmount({ amount: 1000000 });
  iscTx.placeCoinInBag({ bag, coin });

  const targetAddress = agentIdForEVM(
    "0xeC0de100c981B58F8282388cccE340402874bE88"
  );

  iscTx.createAndSend({
    bag,
    contract: getHname("accounts"),
    contractFunction: getHname("transferAllowanceTo"),
    contractArgs: [targetAddress],
    transfers: [[IOTA_TYPE_ARG, 1000000]],
    gasBudget: 10000000n,
  });

  const transaction = iscTx.build();

  // Setting some arbitrary data to make clientless tx building possible.
  transaction.setSender(wallet.toIotaAddress());
  transaction.setGasPayment([]);
  transaction.setGasPrice(1000);
  transaction.setGasBudget(10000000);

  // Excluded passing the client here, to be able to create a transaction with bogus Gas payment data (to estimate without having any funds on L1 or L2)
  const txBytes = await transaction.build();

  return txBytes;
}

async function main() {
  const wallet = Ed25519Keypair.generate();
  const txBytes = await createTestTransaction(wallet);

  estimateOnLedger(txBytes);
}

main();
