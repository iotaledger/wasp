import { GetOwnedObjectsParams, IotaClient, IotaObjectResponse, ObjectOwner } from '@iota/iota-sdk/client';
import { paginatedRequest } from './migration/page_reader';
import { Ed25519Keypair, Ed25519PublicKey } from '@iota/iota-sdk/keypairs/ed25519';
import { toHEX } from '@iota/iota-sdk/utils';

import { input } from '@inquirer/prompts';

import TransportNodeHid from '@ledgerhq/hw-transport-node-hid';
import IOTA_TRANSPORT, { GetPublicKeyResult, GetVersionResult } from '@iota/ledgerjs-hw-app-iota';

import { createMigrationTransaction, MigratedAssets } from './migration/migration_executer';
import { messageWithIntent, toSerializedSignature } from '@iota/iota-sdk/cryptography';

import * as fs from 'fs';
import chalk from 'chalk';
import ora from 'ora';
import { Logger } from './logger';

function newSpinner() {
  return ora({ hideCursor: true, discardStdin: true });
}

// the Mainnet test seed Keypair
// This needs to be replaced with an alternative signing solution

const keypair = Ed25519Keypair.deriveKeypair('gospel poem coffee duty cluster plug turkey buffalo aim annual essay mushroom');
const CONFIG_FILE_PATH = './migration_preparation.json';
const LEDGER_BIP_PATH = "44'/4218'/0'/0'/0'";

interface PrepareConfig {
  CommitteeAddress: string;
  ChainOwner: string;
  AnchorID: string;
  PackageID: string;
  L1ApiUrl: string;
}

async function validateObject(client: IotaClient, name: string, objectID: string, expectedOwner: string): Promise<boolean> {
  const spinner = newSpinner().start(`Validating ${name}: ${objectID}`);

  try {
    const response = await client.getObject({ id: objectID, options: { showOwner: true } });

    if (response.error) {
      spinner.fail(`${name}: Error code: ${response.error.code}`);
      return false;
    }

    if (!response.data?.owner) {
      spinner.fail(`${name}: Failed to get owner of ${name}`);
      return false;
    }

    const objectOwner: ObjectOwner = response.data.owner!;

    if (typeof objectOwner === 'object' && objectOwner !== null && 'AddressOwner' in objectOwner && objectOwner.AddressOwner === expectedOwner) {
      spinner.succeed(`${name}: Validated: ${objectID}`);
      return true;
    }

    spinner.fail(`Invalid owner! ${name} is not owned by your Ledger!`);
    return false;
  } catch (error) {
    spinner.fail(`Error validating ${name}: ${error}`);
    return false;
  }
}

async function readPrepareConfiguration(path: string, expectedOwner: string): Promise<[IotaClient, PrepareConfig, Date]> {
  let prepareConfigStr: string;

  try {
    prepareConfigStr = fs.readFileSync(path, 'utf8');
  } catch (ex) {
    throw new Error(`Failed to read config file: ${ex}`);
  }

  try {
    const config = JSON.parse(prepareConfigStr) as PrepareConfig;
    const stat = fs.statSync(path);

    Logger.info(`Configuration file, modified last: ${stat.mtime}`);
    Logger.info(`Configured Endpoint: ${config.L1ApiUrl}`);

    // Create client with the configured endpoint
    const client = new IotaClient({
      url: config.L1ApiUrl,
    });

    // Validate critical objects
    const anchorValid = await validateObject(client, 'Anchor', config.AnchorID, expectedOwner);

    if (!anchorValid) {
      throw new Error('One or more objects failed validation');
    }

    return [client, config, stat.mtime];
  } catch (error) {
    throw new Error(`Failed to parse or validate config: ${error}`);
  }
}

async function prepareLedger(): Promise<[IOTA_TRANSPORT, string, GetPublicKeyResult]> {
  const spinner = newSpinner().start('Establishing connection to Ledger');

  try {
    const transport = await TransportNodeHid.create();
    const iotaTransport = new IOTA_TRANSPORT(transport);

    // Verify app version
    let version: GetVersionResult;
    try {
      version = await iotaTransport.getVersion();
    } catch (ex) {
      spinner.fail('Connection failed');
      throw new Error('Failed to get the installed app version. Is the Ledger unlocked?');
    }

    if (version.minor != 9) {
      spinner.fail('Incompatible app version');
      throw new Error('Unsupported app version. Are you accidentally using the Stardust IOTA App?');
    }

    spinner.succeed('Connection established');
    Logger.info(`IOTA App validated: ${version.major}.${version.minor}.${version.patch}`);

    // Get public key and address
    let publicKey: GetPublicKeyResult;
    try {
      publicKey = await iotaTransport.getPublicKey(LEDGER_BIP_PATH);
    } catch (ex) {
      throw new Error('Failed to get public key. Is the Ledger unlocked?');
    }

    const address = `0x${toHEX(publicKey.address)}`;

    Logger.info(`Your address is ${address}`);

    return [iotaTransport, address, publicKey];
  } catch (error) {
    throw new Error(`Ledger initialization failed: ${error}`);
  }
}

function displayMigratedAssets(assets: MigratedAssets): void {
  Logger.info(' Native Token:');
  if (Object.keys(assets.NativeToken).length === 0) {
    Logger.info('   None');
  } else {
    for (const [key, value] of Object.entries(assets.NativeToken)) {
      Logger.info(`   ${key}: ${value}`);
    }
  }

  Logger.info(' NFTs:');
  if (Object.keys(assets.NFTs).length === 0) {
    Logger.info('   None');
  } else {
    for (const [key, value] of Object.entries(assets.NFTs)) {
      Logger.info(`   ${key}: ${value}`);
    }
  }

  Logger.info(' Foundries:');
  if (Object.keys(assets.Foundries).length === 0) {
    Logger.info('   None');
  } else {
    for (const [key, value] of Object.entries(assets.Foundries)) {
      Logger.info(`   ${key}: ${value}`);
    }
  }
}

async function main() {
  try {
    Logger.header('L1 STARDUST MIGRATION TOOL');

    Logger.info('This tool will lead you through the migration of a Stardust AliasOutput');
    Logger.warn('The action, once executed, cannot be reverted and can lead to the potential loss of all funds!');
    Logger.info('Until you get asked to sign a transaction, no funds will be moved and the process can be exited at any time.');

    Logger.divider();

    Logger.info(`The tool expects a ${chalk.red("'migration_prepare.json'")} file in your current working directory.`);
    Logger.info('This file contains specific information about where the migrated funds will be moved into.');
    Logger.info("You get this file after invoking 'wasp-cli chain deploy-migrated prepare'");

    Logger.divider();

    Logger.info('Connect your Ledger, unlock it, open the IOTA Rebased App and make sure that Blind Signing is enabled.');

    await input({ message: 'Press enter to proceed', theme: { prefix: '' } });

    // Initialize Ledger
    Logger.header('LEDGER INITIALIZATION');
    const [ledger, address, publicKey] = await prepareLedger();

    // Read configuration
    Logger.header('READING MIGRATION CONFIGURATION');
    const [client, config] = await readPrepareConfiguration(CONFIG_FILE_PATH, address);

    const s = newSpinner();

    s.start('Validating chain owner being equal to the Ledger address');
    if (address != config.ChainOwner) {
      s.fail('Validation failure!');
      throw new Error(`Your ledger address is ${address}, the expected chain owner is: ${config.ChainOwner}`);
    }
    s.succeed("ChainOwner address == Ledger address: validated");

    // Find Stardust AliasOutputs
    Logger.header('STARDUST ALIASOUTPUTS');

    const spinner = newSpinner().start('Searching for AliasOutputs');
    const objects = await paginatedRequest<IotaObjectResponse, GetOwnedObjectsParams>(x => client.getOwnedObjects(x), {
      owner: address,
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
    spinner.succeed(`Found ${aliasObjects.length} AliasOutput(s)`);

    if (aliasObjects.length === 0) {
      throw new Error('No AliasOutputs found. Expected at least 1.');
    }

    Logger.info('Found the following AliasOutputs:');
    for (const obj of aliasObjects) {
      Logger.info(` * ${obj.data?.objectId}`);
    }

    if (aliasObjects.length !== 1) {
      throw new Error(`Invalid number of Alias objects: ${aliasObjects.length}, expected: 1`);
    }

    // Create migration transaction
    const aliasObject = aliasObjects[0];
    const createTxSpinner = newSpinner().start('Creating migration transaction');
    const [aliasOutputConsumeTX, migratedObjects] = await createMigrationTransaction(
      client,
      config.PackageID,
      address,
      aliasObject.data?.objectId!,
      config.AnchorID,
    );
    aliasOutputConsumeTX.setSender(address);
    createTxSpinner.succeed('Migration transaction created');

    // Display migration assets
    Logger.header('TO-BE MIGRATED ASSETS');
    displayMigratedAssets(migratedObjects);

    // Build transaction
    const buildTxSpinner = newSpinner().start('Building unsigned transaction');
    const unsignedTX = await aliasOutputConsumeTX.build({
      client: client,
    });
    buildTxSpinner.succeed('Transaction built successfully');

    await input({ message: 'Press enter to dry-run(test) the migration transaction.' });

    // Dry run
    const dryRunSpinner = newSpinner().start('Performing dry run');
    const dryRun = await client.dryRunTransactionBlock({
      transactionBlock: unsignedTX,
    });

    if (dryRun.effects.status.error || dryRun.effects.status.status == 'failure') {
      dryRunSpinner.fail("Failed to execute dry-run!");
      console.log(dryRun.effects.status);
      throw new Error("Failed to execute dry-run!");
    }
    
    dryRunSpinner.succeed('Dry run completed');

    console.log('Balance changes', dryRun.balanceChanges);
    console.log('Status', dryRun.effects.status);
    console.log('GasUsed', dryRun.effects.gasUsed);

    fs.writeFileSync("dry-run.json", JSON.stringify(dryRun, null, 4));

    Logger.warn("a 'dry-run.json' file has been created in your current working directory. It contains the full output of the dry run. Take a good look.\n");

    // Execute transaction
    await input({ message: 'Press enter to execute the transaction. A signature request will be displayed on your Ledger. After you sign this transaction, it will be posted to L1.' });

    const executeTxSpinner = newSpinner().start('Preparing transaction for signing');
    const txBytes = await aliasOutputConsumeTX.build({ client: client });

    const msg = messageWithIntent('TransactionData', txBytes);
    executeTxSpinner.text = 'Waiting for Ledger signature...';

    const signedBlock = await ledger.signTransaction(LEDGER_BIP_PATH, msg);
    const signed = toSerializedSignature({
      signatureScheme: 'ED25519',
      signature: signedBlock.signature,
      publicKey: new Ed25519PublicKey(publicKey.publicKey),
    });

    executeTxSpinner.succeed('Transaction signed successfully');

    const submitTxSpinner = newSpinner().start('Submitting transaction');
    const result = await client.executeTransactionBlock({
      signature: signed,
      transactionBlock: txBytes,
    });
    submitTxSpinner.succeed('Transaction submitted successfully');

    Logger.header('TRANSACTION RESULT');
    console.log(result);

    Logger.success('Migration completed successfully!');
    process.exit(0);
  } catch (error) {
    Logger.error(`Migration failed: ${error}`);
    process.exit(1);
  }
}

main();
