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
  tx.setGasBudget(50000000000);

  const [anchor, stateController, governance] = tx.moveCall({
    target: `${contractId}::anchor::start_new_chain`,
  });

  tx.transferObjects([anchor, stateController, governance], tx.pure(keyPair.toSuiAddress()));

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


  if (result.effects?.status.status === "failure" || result.effects?.status.error) {
    throw new Error(result.effects.status.error);
  }

  console.log("New chain started:");

  return result;
}