'use strict';

import { SuiClient, } from '@mysten/sui.js/client';
import { Ed25519Keypair } from '@mysten/sui.js/keypairs/ed25519';
import { TransactionBlock } from '@mysten/sui.js/transactions';
import { prettyPrint } from './utils';
import { exec, ExecException } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export interface ContractBlobData {
  modules: string[];
  dependencies: string[];
  digest: number[];
}

export async function buildMoveContract(baseDir: string): Promise<ContractBlobData> {
  const { stdout, stderr } = await execAsync('sui move build --dump-bytecode-as-base64', {
    cwd: baseDir,
  });

  /* stderr is used by sui to output compile info such as "including dependency", not only to output real error messages. */
  if (stderr.includes("ERR")) {
    throw new Error(stderr);
  }

  const blobData: ContractBlobData = JSON.parse(stdout);

  if (!blobData.modules || !blobData.dependencies) {
    throw new Error(`failed to parse contract blob ${stderr}`);
  }

  if (blobData.modules.length == 0 || blobData.dependencies.length == 0) {
    throw new Error('contract does not seem to contain any modules or dependencies');
  }

  return blobData;

}

export async function publishContract(client: SuiClient, keyPair: Ed25519Keypair, baseDir: string) {
  const tx = new TransactionBlock();
  tx.setGasBudget(50000000000);

  const blob = await buildMoveContract(baseDir);

  const pub = tx.publish({
    modules: [...blob.modules],
    dependencies: blob.dependencies
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
    prettyPrint(block);
    throw new Error(block.effects.status.error);
  }

  return block;
}
