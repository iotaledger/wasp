import { GetOwnedObjectsParams, IotaClient, IotaObjectResponse, IotaTransport, ObjectOwner } from '@iota/iota-sdk/client';
import { paginatedRequest } from './page_reader';
import { Ed25519Keypair, Ed25519PublicKey } from '@iota/iota-sdk/keypairs/ed25519';
import { toB64, toHEX } from '@iota/iota-sdk/utils';

import { input, confirm, select, checkbox } from '@inquirer/prompts';

import TransportNodeHid from '@ledgerhq/hw-transport-node-hid';
import IOTA_TRANSPORT, { GetPublicKeyResult, GetVersionResult } from '@iota/ledgerjs-hw-app-iota';

import { createMigrationTransaction } from './migration_executer';

import * as fs from 'fs';
import * as util from 'util';
import chalk from 'chalk';
import ora from 'ora';
import { messageWithIntent, PublicKey, toSerializedSignature } from '@iota/iota-sdk/cryptography';
util.inspect.defaultOptions.depth = null;

function newSpinner() {
  return ora({ hideCursor: true, discardStdin: true });
}

// the Mainnet test seed Keypair
// This needs to be replaced with an alternative signing solution

const GOVERNOR_ADDRESS = '0x70bc12d8964837afac5978b4e3acc61defe9427e0c975afb1f3663c186e3b1e6';
const keypair = Ed25519Keypair.deriveKeypair('gospel poem coffee duty cluster plug turkey buffalo aim annual essay mushroom');

const LEDGER_BIP_PATH = "44'/4218'/0'/0'/0'";

interface PrepareConfig {
  CommitteeAddress: string;
  ChainOwner: string;
  AssetsBagID: string;
  GasCoinID: string;
  AnchorID: string;
  PackageID: string;
  L1ApiUrl: string;
}

async function validateObject(client: IotaClient, name: string, objectID: string, expectedOwner: string) {
  let s = newSpinner().start(`Validating ${name}: ${objectID}`);
  const gasCoin = await client.getObject({ id: objectID, options: { showOwner: true } });

  if (gasCoin.error) {
    s.fail(gasCoin.error.code);
    return;
  }

  if (!gasCoin.data?.owner) {
    s.fail(`Failed to get owner of ${name}`);
    return;
  }

  const objectOwner: ObjectOwner = gasCoin.data.owner!;

  if (typeof objectOwner == 'object' && objectOwner != null && 'AddressOwner' in objectOwner && objectOwner.AddressOwner == expectedOwner) {
    s.succeed(`${name} Validated: ${objectID}`);
    return;
  }

  s.fail(`Invalid owner! ${name} is not owned by your Ledger!`);
}

async function readPrepareConfiguration(path: string, expectedOwner: string): Promise<[client: IotaClient, config: PrepareConfig, modifiedAt: Date]> {
  let prepareConfigStr: string;

  try {
    prepareConfigStr = fs.readFileSync(path, 'utf8');
  } catch (ex) {
    throw new Error(`Failed to read config file: ${ex}`);
  }

  const config = JSON.parse(prepareConfigStr) as PrepareConfig;
  const stat = fs.statSync(path);

  // Validate that all required IDs are available for use.

  const client = new IotaClient({
    url: config.L1ApiUrl,
  });

  console.log(` Configuration file, modified last: ${stat.mtime}`);
  console.log(' Configured Endpoint: ' + config.L1ApiUrl);

  await validateObject(client, 'Gas Coin', config.GasCoinID, expectedOwner);
  await validateObject(client, 'Anchor', config.AnchorID, expectedOwner);
  await validateObject(client, 'AssetsBag', config.AssetsBagID, expectedOwner);

  return [client, config, stat.mtime];
}

async function prepareLedger(): Promise<[transport: IOTA_TRANSPORT, address: string, publicKey: GetPublicKeyResult]> {
  let s = newSpinner().start('Establishing connection');

  const transport = await TransportNodeHid.create();
  const iotaTransport = new IOTA_TRANSPORT(transport);

  let version: GetVersionResult;
  let publicKey: GetPublicKeyResult;

  try {
    version = await iotaTransport.getVersion();
  } catch (ex) {
    throw new Error('Failed to get the installed app version. Is the Ledger unlocked?');
  }

  if (version.minor != 9) {
    throw new Error('Unsupported app version. Are you accidentially using the Stardust IOTA App?');
  }

  s.succeed('Connection established');

  console.log(` IOTA App validated: ${version.major}.${version.minor}.${version.patch}`);

  try {
    publicKey = await iotaTransport.getPublicKey(LEDGER_BIP_PATH);
  } catch (ex) {
    throw new Error('Failed to get the installed app version. Is the Ledger unlocked?');
  }

  const address = `0x${toHEX(publicKey.address)}`;

  console.log(` Your address is ${address}`);

  return [iotaTransport, address, publicKey];
}

async function main() {
  console.log(chalk.red('L1 Startdust migration tool\n'));
  console.log(chalk.white('This tool will lead you through the migration of a Stardust AliasOutput'));
  console.log(chalk.white('The action, once executed, can not be reverted and can lead to the potential ') + chalk.red('loss of all funds!'));
  console.log(chalk.white('Until you get asked to sign a transaction, no funds will be moved and the process can be exited at any time.'));

  console.log('\n\n---------------------------------------------------------\n\n');

  console.log(
    `The tool expects a ${chalk.red("'migration_prepare.json'")} file in your current working directory.\n` +
      "This file contains specific information about where the migrated funds will be moved into.\nYou get this file after invoking 'wasp-cli chain deploy-migrated prepare'\n",
  );

  console.log('\n\n---------------------------------------------------------\n\n');

  console.log(`Connect your Ledger, unlock it, open the IOTA Rebased App and make sure that Blind Signing is enabled.`);

  await input({ message: 'Press enter to proceed', theme: { prefix: '' } });

  console.log(chalk.cyan('\nLedger initialization\n'));

  const [ledger, address, publicKey] = await prepareLedger();

  console.log(chalk.cyan('\nReading Migration configuration\n'));

  const [client, config] = await readPrepareConfiguration('./migration_preparation.json', address);

  console.log(chalk.cyan('\nStardust AliasOutputs\n'));

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

  if (aliasObjects.length == 0) {
    throw new Error(chalk.red(`Invalid amount of Alias objects: ${aliasObjects.length}, expected: 1`));
  }

  console.log('Found the following AliasOutputs:');

  for (let o of aliasObjects) {
    console.log(` * ${o.data?.objectId}`);
  }

  if (aliasObjects.length != 1) {
    throw new Error(chalk.red(`Invalid amount of Alias objects: ${aliasObjects.length}, expected: 1`));
  }

  const aliasObject = aliasObjects[0];
  const [aliasOutputConsumeTX, migratedObjects] = await createMigrationTransaction(
    client,
    config.PackageID,
    GOVERNOR_ADDRESS,
    aliasObject.data?.objectId!,
    config.AnchorID,
  );
  aliasOutputConsumeTX.setSender(GOVERNOR_ADDRESS);

  console.log(chalk.cyan('\nTo-be migrated assets\n'));

  console.log(' Native Token:');

  for (let n of Object.keys(migratedObjects.nativeToken)) {
    console.log(`   ${n}: ${migratedObjects.nativeToken[n]}`);
  }

  console.log(' NFTs:');

  for (let n of Object.keys(migratedObjects.NFTs)) {
    console.log(`   ${n}: ${migratedObjects.NFTs[n]}`);
  }

  console.log(' Foundries:');

  for (let n of Object.keys(migratedObjects.Foundries)) {
    console.log(`   ${n}: ${migratedObjects.Foundries[n]}`);
  }

  console.log('Unsigned Tx:');

  const unsignedTX = await aliasOutputConsumeTX.build({
    client: client,
  });

  await input({ message: 'Press enter to show the result of the dry-run' });

  const dryRun = await client.dryRunTransactionBlock({
    transactionBlock: unsignedTX,
  });

  console.log(dryRun);

  await input({ message: 'Press enter to execute the transaction. A signage request will be displayed on your Ledger.' });

  const txBytes = await aliasOutputConsumeTX.build({ client: client });

  const msg = messageWithIntent('TransactionData', txBytes);
  const signedBlock = await ledger.signTransaction(LEDGER_BIP_PATH, msg);
  const signed = toSerializedSignature({
    signatureScheme: 'ED25519',
    signature: signedBlock.signature,
    publicKey: new Ed25519PublicKey(publicKey.publicKey),
  });

  const result = await client.executeTransactionBlock({
    signature: signed,
    transactionBlock: txBytes,
  });

  console.log(result);
}

main();
