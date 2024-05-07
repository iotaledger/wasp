import { SuiClient, SuiObjectChangePublished, type SuiObjectChange, type SuiObjectChangeCreated } from "@mysten/sui.js/client";
import { toB64 } from '@mysten/sui.js/utils';
import { bcs } from "@mysten/sui.js/bcs";
import { Ed25519Keypair } from "@mysten/sui.js/keypairs/ed25519";
import { TransactionBlock } from "@mysten/sui.js/transactions";
import { prettyPrint } from "../utils";
import { publishContract } from "../contract_builder";
import { BCS, fromB58, fromB64, getSuiMoveConfig } from "@mysten/bcs";

export interface ISCChainInfo {
  packageId: string;
  anchorCapId: string;
}

export class Command {
  targetContract: string;
  entryPoint: string;
  params: number[][];

  constructor(targetContract: string, entryPoint: string, params: number[][]) {
    this.targetContract = targetContract;
    this.entryPoint = entryPoint;
    this.params = params;
  }
}

const Balance = bcs.struct('Balance', {
  value: bcs.u64(),
}, {});

const BCSCommand = bcs.struct('Command', {
  targetContract: bcs.string(),
  entryPoint: bcs.string(),
  params: bcs.vector(bcs.vector(bcs.u8())),
});

export class ChainStructure {

  anchorId: string;
  governorId: string;

  constructor(anchorId: string, governanceId: string) {
    this.anchorId = anchorId;
    this.governorId = governanceId;
  }
}

export class ISCAnchorContract {

  client: SuiClient;
  keyPair: Ed25519Keypair;
  iscChainInfo: ISCChainInfo;

  constructor(client: SuiClient, keyPair: Ed25519Keypair, iscChainInfo: ISCChainInfo) {
    this.client = client;
    this.keyPair = keyPair;
    this.iscChainInfo = iscChainInfo;
  }

  async start_new_chain(): Promise<ChainStructure> {
    const tx = new TransactionBlock();
    tx.setGasBudget(50000000000);

    const [governorCap] = tx.moveCall({
      target: `${this.iscChainInfo.packageId}::anchor::start_new_chain`,
      arguments: [tx.pure(this.iscChainInfo.anchorCapId)]
    });

    tx.transferObjects([governorCap], tx.pure(this.keyPair.toSuiAddress()));

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

    if (result.effects?.created?.length != 2) {
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

    const findAndGetObjId = (name: string) => {
      const item = result.objectChanges?.find(x => x.type === 'created' && x.objectType.endsWith(name));

      if (!item) {
        throw new Error(`Can not find ${name} in objectChanges`);
      }

      return item['objectId'];
    }

    const anchorObj = findAndGetObjId('Anchor');
    const governorObj = findAndGetObjId('GovernorCap');

    console.log("Anchor", anchorObj, "governance", governorObj);

    return new ChainStructure(anchorObj, governorObj)
  }

  // TODO: This vector<vector<u8>> serialization does not work and throws errors.
  serializeVectorOfVectors(data: Uint8Array[]): string {
    const serialized = data.map(innerVector => {
      const innerSerialized = toB64(innerVector);
      return innerSerialized.toString();
    });
    return JSON.stringify(serialized);
  }

  async send_request(anchorId: string, command: Command, value: BigInt = 0n, asset: BigInt | null = null) {
    const tx = new TransactionBlock();
    tx.setGasBudget(50000000000);

    const d = this.serializeVectorOfVectors([new Uint8Array([0, 1, 2, 3]), new Uint8Array([0]),])

    const [request] = tx.moveCall({
      arguments: [
        tx.pure.string(command.targetContract),
        tx.pure.string(command.entryPoint),
        tx.pure(d),
      ],
      target: `${this.iscChainInfo.packageId}::request::create_request`
    })

    const [res] = tx.moveCall({
      arguments: [
        tx.pure(anchorId),
        request,
      ],
      target: `${this.iscChainInfo.packageId}::anchor::send_request`,
      //  typeArguments: ['0x2::sui::SUI']
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



export async function publishISCContract(client: SuiClient, keyPair: Ed25519Keypair, basePath: string): Promise<ISCChainInfo> {
  console.log('Publishing ISC contract');

  // It is expected to have the kinesis:isc-models repo next to the isc-private repo.
  const block = await publishContract(client, keyPair, basePath);

  const publishedPackage = block.objectChanges?.find(x => x.type === 'published');

  if (!publishedPackage) {
    throw new Error('Can not find packageId');
  }

  const anchorCap = block.objectChanges?.find(x => x.type == 'created' && x.objectType.endsWith('AnchorCap'));

  if (!anchorCap) {
    throw new Error('Did not receive AnchorCap');
  }

  const pubPackage = publishedPackage as SuiObjectChangePublished;
  console.log(`Success. PackageId: ${pubPackage.packageId}`);

  return {
    packageId: pubPackage.packageId,
    anchorCapId: (anchorCap as SuiObjectChangeCreated).objectId,
  };
}