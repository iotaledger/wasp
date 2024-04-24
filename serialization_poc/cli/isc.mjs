import { SuiClient } from "@mysten/sui.js/client";
import { Ed25519Keypair } from "@mysten/sui.js/keypairs/ed25519";
import { TransactionBlock } from "@mysten/sui.js/transactions";

export class Request { }

/**
 * 
 * @param {SuiClient} client 
 * @param {Ed25519Keypair} keyPair 
 * @param {string} contractId
 */
export async function start_new_chain(client, keyPair, contractId) {
  const tx = new TransactionBlock();

  const chainCall = tx.moveCall({
    arguments: [],
    target: `${contractId}::anchor::start_new_chain`,

  });

  tx.setGasBudget(50000000000);
  tx.transferObjects([chainCall], tx.pure(keyPair.toSuiAddress()));

  const result = await client.signAndExecuteTransactionBlock({
    transactionBlock: tx,
    signer: keyPair,
    options: {
      showEffects: true,
      showObjectChanges: true,
      showEvents: true,
    }
  });

  await client.waitForTransactionBlock({
    digest: result.digest,
  })

  return result;
}