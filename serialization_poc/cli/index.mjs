'use strict';

import { requestSuiFromFaucetV1 } from '@mysten/sui.js/faucet';
import { getFullnodeUrl, SuiClient } from '@mysten/sui.js/client';
import { MIST_PER_SUI } from '@mysten/sui.js/utils';
import { Ed25519Keypair } from '@mysten/sui.js/keypairs/ed25519';
import { TransactionBlock } from '@mysten/sui.js/transactions';
import { readFileSync } from 'fs';
import { start_new_chain } from './isc.mjs';

const SUI_HOST = "http://127.0.0.1:9000";
const SUI_FAUCET_HOST = "http://127.0.0.1:9123";
const SUI_BASE = BigInt(1000000000000);

const KEYPAIR_SECRET = new Uint8Array(32); // zero'd array with a length of 32 to get the same KeyPair every time.
const keypair = Ed25519Keypair.fromSecretKey(KEYPAIR_SECRET);
const address = keypair.toSuiAddress();

async function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function parallelFaucetRequest(nTimes, address) {
  const requests = Array.from({ length: nTimes }, () =>
    requestSuiFromFaucetV1({
      host: SUI_FAUCET_HOST,
      recipient: address,
    })
  );

  await Promise.all(requests);
}

/**
 * 
 * @param {SuiClient} client 
 * @param {Ed25519Keypair} keypair
 * @param {string} contractPath
 */
async function lazyPublish(client, keypair, contractPath) {
  const contract = readFileSync(contractPath);
  const tx = new TransactionBlock();

  const pub = tx.publish({
    modules: [[...contract]],
    dependencies: [
      "0x0000000000000000000000000000000000000000000000000000000000000001",
      "0x0000000000000000000000000000000000000000000000000000000000000002"
    ]
  });

  tx.transferObjects([pub], tx.pure(keypair.toSuiAddress()));

  const block = await client.signAndExecuteTransactionBlock({
    transactionBlock: tx,
    signer: keypair,

    options: {
      showEffects: true,
      showObjectChanges: true,
    },
  });

  await client.waitForTransactionBlock({ digest: block.digest });

  return block;
}

const CONTRACT_ID = '0xeb5c15754765cb7f3da16b37a79d93a2049299f9b42d22477f35ec0169274aa7';

/**
 * 
 * @param {SuiClient} client 
 * @param {Ed25519Keypair} keyPair 
 */
async function handleFunds(client, keyPair) {
  const balance = await client.getBalance({
    owner: keyPair.toSuiAddress(),
  });

  if (BigInt(balance.totalBalance) < SUI_BASE) {
    await parallelFaucetRequest(10, address);
    await delay(15000);
  }
}

/**
 * 
 * @param {SuiClient} client 
 * @param {Ed25519Keypair} keyPair 
 */
async function handlePublishContract(client, keyPair) {
  const block = await lazyPublish(client, keypair, '/home/luke/dev/isc-private/serialization_poc/contract/isc/build/isc/bytecode_modules/anchor.mv');
  const publishedPackage = block.objectChanges.find(x => x.type === 'published');

  console.log(publishedPackage.packageId);
  return publishedPackage.packageId;
}

async function main() {
  const keyPair = Ed25519Keypair.fromSecretKey(KEYPAIR_SECRET);
  const client = new SuiClient({ url: SUI_HOST });

  // await handleFunds(client, keyPair);
  // await handlePublishContract(client, keyPair);

  const chain = await start_new_chain(client, keyPair, CONTRACT_ID);

  console.log(chain);
}


main();