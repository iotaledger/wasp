import { SuiClient } from "@mysten/sui.js/client";
import { bcs } from "@mysten/sui.js/bcs";
import { Ed25519Keypair } from "@mysten/sui.js/keypairs/ed25519";
import { TransactionBlock } from "@mysten/sui.js/transactions";
import { prettyPrint } from "./utils.mjs";

export class Command {
  targetContract = 0;
  entryPoint = 0;
  params;
  allowance;
  gasBudget = 0;


  /**
   * @param {number} targetContract
   * @param {number} entryPoint
   * @param {number[][]} params
   * @param {*} allowance
   * @param {number} gasBudget
   */
  constructor(targetContract, entryPoint, params, allowance, gasBudget) {
    this.targetContract = targetContract;
    this.entryPoint = entryPoint;
    this.params = params;
    this.allowance = allowance;
    this.gasBudget = gasBudget;
  }
}

const Balance = bcs.struct('Balance', {
  value: bcs.u64(),
}, {});

const BCSCommand = bcs.struct('Command', {
  targetContract: bcs.u64(),
  entryPoint: bcs.u64(),
  params: bcs.vector(bcs.vector(bcs.u8())),
  allowance: bcs.option(Balance),
  gasBudget: bcs.u64(),
});



export class ChainStructure {
  /**
   * @param {string} anchorId
   * @param {string} stateControllerId
   * @param {string} governanceId
   */
  constructor(anchorId, stateControllerId, governanceId) {
    this.anchorId = anchorId;
    this.stateControllerId = stateControllerId;
    this.governorId = governanceId;
  }
}

export class ISCAnchorContract {
  /**
   * @param {SuiClient} client
   * @param {Ed25519Keypair} keyPair
   * @param {string} packageId
   */
  constructor(client, keyPair, packageId) {
    this.client = client;
    this.keyPair = keyPair;
    this.packageId = packageId;
  }


  /**
   * @returns {Promise<ChainStructure>}
   */
  async start_new_chain() {
    const tx = new TransactionBlock();
    tx.setGasBudget(50000000000);

    const [anchor, stateController, governor] = tx.moveCall({
      target: `${this.packageId}::anchor::start_new_chain`,
    });

    tx.transferObjects([anchor, stateController, governor], tx.pure(this.keyPair.toSuiAddress()));

    const result = await this.client.signAndExecuteTransactionBlock({
      transactionBlock: tx,
      signer: this.keyPair,

      options: {
        showEffects: true,
        showObjectChanges: true,
        showEvents: true,

      }
    });

    await this.client.waitForTransactionBlock({
      digest: result.digest,
    })


    if (result.effects?.status.status === "failure" || result.effects?.status.error) {
      throw new Error(result.effects.status.error);
    }

    console.log("New chain started:");

    if (result.effects?.created?.length != 3) {
      throw new Error("Unexpected return result from start_new_chain");
    }

    /**
     * Initially I've thought something like this would work to select all three return types:
     * const [anchorObj, stateControllerObj, governorObj] = [...result.effects.created.map(x => x.reference)];
     * 
     * As Anchor, StateController, Governance is returned in this order from Move.
     * But StateController was on index 0 and anchor on index 2.
     * For now selecting them manually, until I find a better solution 
     */

    const findAndGetObjId = (name) => {
      const item = result.objectChanges?.find(x => x.type === 'created' && x.objectType.endsWith(name));

      if (!item) {
        throw new Error(`Can not find ${name} in objectChanges`);
      }

      // @ts-ignore
      return item['objectId']; // This is stupid, might need to switch TypeScript or find a better approach. The property is there, just hidden in JSDoc. :D
    }

    const anchorObj = findAndGetObjId('Anchor');
    const stateControllerObj = findAndGetObjId('StateControllerCap');
    const governorObj = findAndGetObjId('GovernorCap');

    console.log("Anchor", anchorObj, "StateController", stateControllerObj, "governance", governorObj);

    return new ChainStructure(anchorObj, stateControllerObj, governorObj)
  }

  /**
   * @param {string} anchorId
   * @param {Command} command
   * @param {BigInt} value
   * @param {BigInt | null} asset
   */
  async send_request(anchorId, command, value = 0n, asset = null) {
    const tx = new TransactionBlock();
    tx.setGasBudget(50000000000);

    let serializedAssetBalance = bcs.option(Balance).serialize(null);
    if (asset) {
      serializedAssetBalance = bcs.option(Balance).serialize({ value: asset.toString() });
    }

    const baseSuiToken = Balance.serialize({ value: value.toString() });

    const serializedParams = bcs.ser("vector<vector<u8>>", command.params);

    const [commandObj] = tx.moveCall({
      arguments: [
        tx.pure(command.targetContract),
        tx.pure(command.entryPoint),
        tx.pure(serializedParams),
        tx.pure(null),
        tx.pure(command.gasBudget),
      ],
      target: `${this.packageId}::anchor::create_command`

    })



    const [res] = tx.moveCall({
      arguments: [
        tx.object(anchorId),
        commandObj,
        serializedAssetBalance,
        baseSuiToken,
      ],
      target: `${this.packageId}::anchor::send_request`,
      typeArguments: ['0x2::sui::SUI']
    });

    tx.transferObjects([res], tx.pure(this.keyPair.toSuiAddress()));

    prettyPrint(JSON.parse(tx.serialize()));

    const result = await this.client.signAndExecuteTransactionBlock({
      transactionBlock: tx,
      signer: this.keyPair,
      options: {
        showEffects: true,
        showObjectChanges: true,
        showEvents: true,
      }
    });

    await this.client.waitForTransactionBlock({
      digest: result.digest,
    })

    return result;
  }
}


