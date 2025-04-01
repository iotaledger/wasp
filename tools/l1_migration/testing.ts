import { IotaClient, IotaObjectChangeCreated } from '@iota/iota-sdk/client';
import { ISCMove } from './migration/isc';
import { Keypair } from '@iota/iota-sdk/cryptography';
import { Transaction } from '@iota/iota-sdk/transactions';

async function createTestAnchor(client: IotaClient, packageId: string, keypair: Keypair) {
  const tx = new Transaction();

  ISCMove.newAnchor(packageId, tx, keypair.toIotaAddress());

  const req = await client.signAndExecuteTransaction({
    transaction: tx,
    signer: keypair,
    options: {
      showObjectChanges: true,
    },
  });

  const result = await client.waitForTransaction({
    digest: req.digest,
    options: {
      showObjectChanges: true,
    },
  });

  const objectId = `${packageId}::anchor::Anchor`;

  const anchor = result.objectChanges!.find(x => x.type == 'created' && x.objectType == objectId);

  return (anchor as IotaObjectChangeCreated).objectId;
}

async function dumpAssetsBag(client: IotaClient) {
  const fields = await client.getDynamicFields({
    parentId: '0x1c76d5d3673c5c7d16dc2f5071502fa1fff32704a36b02207583ee8e81b3e025',
  });

  console.dir(fields, { depth: null });
}
