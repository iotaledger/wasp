'use strict';

import { requestSuiFromFaucetV1 } from '@mysten/sui.js/faucet';
import { getFullnodeUrl, SuiClient } from '@mysten/sui.js/client';
import { MIST_PER_SUI, SUI_DECIMALS } from '@mysten/sui.js/utils';
import { Ed25519Keypair } from '@mysten/sui.js/keypairs/ed25519';
import { TransactionBlock } from '@mysten/sui.js/transactions';
import { readFileSync } from 'fs';

import { start_new_chain } from './isc.mjs';
import { delay, prettyPrint } from './utils.mjs';

const SUI_HOST = "http://127.0.0.1:9000";
const SUI_FAUCET_HOST = "http://127.0.0.1:9123";
const KEYPAIR_SECRET = new Uint8Array(32); // zero'd array with a length of 32 to get the same KeyPair every time.

/**
 * 
 * @param {number} nTimes 
 * @param {string} address 
 */
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
 * @param {Ed25519Keypair} keyPair 
 */
async function handleFunds(client, keyPair) {
  const balance = await client.getBalance({
    owner: keyPair.toSuiAddress(),
  });

  if (BigInt(balance.totalBalance) < MIST_PER_SUI * 10n) {
    const faucetIterations = 10;

    console.log(`Requesting funds ${faucetIterations} times.`);
    await parallelFaucetRequest(faucetIterations, keyPair.toSuiAddress());

    console.log('Waiting for funds to arrive');
    await delay(15000);
  }
}

/**
 * 
 * @param {SuiClient} client 
 * @param {Ed25519Keypair} keyPair
 * @param {string} contractPath
 */
async function publishContractBlob(client, keyPair, contractPath) {
  const contract = readFileSync(contractPath);

  const tx = new TransactionBlock();
  tx.setGasBudget(50000000000);

  const pub = tx.publish({
    modules: [[...contract]],
    dependencies: [
      "0x0000000000000000000000000000000000000000000000000000000000000001",
      "0x0000000000000000000000000000000000000000000000000000000000000002"
    ]
  });

  tx.transferObjects([pub], tx.pure(keyPair.toSuiAddress()));

  const block = await client.signAndExecuteTransactionBlock({
    transactionBlock: tx,
    signer: keyPair,
    options: {
      showEffects: true,
      showObjectChanges: true,
    },
  });

  await client.waitForTransactionBlock({ digest: block.digest });

  if (block.effects?.status.status === 'failure' || block.effects?.status.error) {
    throw new Error(block.effects.status.error);
  }

  return block;
}

/**
 * 
 * @param {SuiClient} client 
 * @param {Ed25519Keypair} keyPair 
 */
async function handlePublishContract(client, keyPair) {
  console.log('Publishing anchor contract');

  const block = await publishContractBlob(client, keyPair, '/home/luke/dev/isc-private/serialization_poc/contract/isc/build/isc/bytecode_modules/anchor.mv');
  const publishedPackage = block.objectChanges?.find(x => x.type === 'published');

  if (!publishedPackage) {
    throw new Error('Can not find packageId');
  }

  console.log(`Success. PackageId: ${publishedPackage.packageId}`);

  return publishedPackage.packageId;
}

async function main() {
  const keyPair = Ed25519Keypair.fromSecretKey(KEYPAIR_SECRET);
  const client = new SuiClient({ url: SUI_HOST });

  await handleFunds(client, keyPair);
  const packageId = await handlePublishContract(client, keyPair);
  const chain = await start_new_chain(client, keyPair, packageId);

  prettyPrint(chain);
}


main();