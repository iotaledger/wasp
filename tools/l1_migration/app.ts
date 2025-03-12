import { GetOwnedObjectsParams, IotaClient, IotaObjectChangeCreated, IotaObjectResponse } from '@iota/iota-sdk/client';
import { paginatedRequest } from './page_reader';
import { Ed25519Keypair } from '@iota/iota-sdk/keypairs/ed25519';
import { toHEX } from '@iota/iota-sdk/utils';

import { executeMigration } from './migration_executer';
import { ISCMove } from './isc';
import { Transaction } from '@iota/iota-sdk/transactions';

import * as util from 'util';

util.inspect.defaultOptions.depth = null

const ENDPOINT_URL = 'http://localhost:9000';
const ISC_PACKAGE_ID = '0xecffc8e3606a4a97044e53a58d8181b69316dc07bb6acdaee8cfb3f64246cbd7';
const GOVERNOR_ADDRESS = '0x70bc12d8964837afac5978b4e3acc61defe9427e0c975afb1f3663c186e3b1e6';

// the Mainnet test seed Keypair
const keypair = Ed25519Keypair.deriveKeypair('gospel poem coffee duty cluster plug turkey buffalo aim annual essay mushroom');

//
// Magic, don't touch.
//
const client = new IotaClient({
  url: ENDPOINT_URL,
});

async function createTestAnchor(packageId: string) {
  const tx = new Transaction();
  
  ISCMove.newAnchor(packageId, tx, keypair.toIotaAddress());

  const req = await client.signAndExecuteTransaction({
    transaction: tx,
    signer: keypair,
    options: {
      showObjectChanges: true,
    }
  });

  const result = await client.waitForTransaction({
    digest: req.digest,
    options: {
      showObjectChanges: true,
    },
  });

  const objectId = `${packageId}::anchor::Anchor`;

  const anchor = result.objectChanges!
    .find(x=>x.type == 'created' && x.objectType == objectId)
    
  return (anchor as IotaObjectChangeCreated).objectId;
}

async function dumpAssetsBag() {
  const fields = await client.getDynamicFields({
    parentId: '0x1c76d5d3673c5c7d16dc2f5071502fa1fff32704a36b02207583ee8e81b3e025'
  });

  console.dir(fields, {depth:null});
}

async function main() {
  await dumpAssetsBag();
  return;
  const anchorObject = await createTestAnchor(ISC_PACKAGE_ID);

  const objects = await paginatedRequest<IotaObjectResponse, GetOwnedObjectsParams>(x => client.getOwnedObjects(x), {
    owner: GOVERNOR_ADDRESS,
    filter: {
      MatchAll: [
        {
          StructType: '0x107a::alias_output::AliasOutput<0x2::iota::IOTA>',
        },
      ],
    },
    options: {
      showType: true,
      showContent: true,
    },
  });

  const aliasObjects = objects.filter(x => x.data?.type == '0x107a::alias_output::AliasOutput<0x2::iota::IOTA>');

  if (aliasObjects.length != 1) {
    throw new Error(`Invalid amount of Alias objects: ${aliasObjects.length}, expected: 1`);
  }

  const aliasObject = aliasObjects[0];
  const aliasOutputConsumeTX = await executeMigration(client, ISC_PACKAGE_ID, GOVERNOR_ADDRESS, aliasObject.data?.objectId!, anchorObject);

  aliasOutputConsumeTX.setSender(GOVERNOR_ADDRESS);
  console.log("Unsigned Tx:");

  const unsignedTX = await aliasOutputConsumeTX.build({
    client: client
  });

  console.log(toHEX(unsignedTX));

  const dryRun = await client.dryRunTransactionBlock({
    transactionBlock: unsignedTX,

  });
  
  console.log(dryRun);
  const result = await client.signAndExecuteTransaction({
    transaction: aliasOutputConsumeTX,
    signer: keypair,
  });

  console.log(result);
}

main();
