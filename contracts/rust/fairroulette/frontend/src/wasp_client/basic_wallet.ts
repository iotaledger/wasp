import { Base58 } from './crypto';
import { Colors } from './colors';
import type { BasicClient } from './basic_client';
import type { IWalletAddressOutput, IWalletOutput, IWalletOutputBalance } from './models';

export class BasicWallet {

  private client: BasicClient;
  constructor(client: BasicClient) {
    this.client = client;
  }

  private mapToArray(b: { [color: string]: bigint; }): (IWalletOutputBalance[]) {
    const balances: IWalletOutputBalance[] = [];

    for (const [color, value] of Object.entries(b)) {
      let colorName = color;
      if (color === Base58.encode(Colors.IOTA_COLOR_BYTES)) {
        colorName = Colors.IOTA_COLOR_STRING;
      }
      balances.push({
        color: colorName,
        value: BigInt(value)
      });
    }
    return balances;
  }

  public async getUnspentOutputs(address: string) {
    const unspents = await this.client.unspentOutputs({ addresses: [address] });

    const usedAddresses = unspents.unspentOutputs.filter(u => u.outputs.length > 0);

    const unspentOutputs = usedAddresses.map(uo => ({
      address: uo.address.base58,
      outputs: uo.outputs.map(uid => ({
        id: uid.output.outputID.base58,
        balances: this.mapToArray(uid.output.output.balances),
        inclusionState: uid.inclusionState
      }))
    }));

    return unspentOutputs;
  }

  public determineOutputsToConsume(unspentOutputs: IWalletAddressOutput[], destinationAddress: string, iotas: bigint): {
    [address: string]: { [outputID: string]: IWalletOutput; };
  } {
    const outputsToConsume: { [address: string]: { [outputID: string]: IWalletOutput; }; } = {};

    let iotasLeft = iotas;

    for (const unspentOutput of unspentOutputs) {
      let outputsFromAddressSpent = false;

      const confirmedUnspentOutputs = unspentOutput.outputs.filter(o => o.inclusionState.confirmed);

      for (const output of confirmedUnspentOutputs) {
        let requiredColorFoundInOutput = false;

        const balance = output.balances.find(x => x.color == Colors.IOTA_COLOR_STRING);

        if (!balance) {
          continue;
        }

        if (iotasLeft > 0n) {
          if (iotasLeft > balance.value) {
            iotasLeft -= balance.value;
          } else {
            iotasLeft = 0n;
          }

          requiredColorFoundInOutput = true;
        }

        // if we found required tokens in this output
        if (requiredColorFoundInOutput) {
          // store the output in the outputs to use for the transfer
          outputsToConsume[unspentOutput.address] = {};
          outputsToConsume[unspentOutput.address][output.id] = output;

          // mark address as spent
          outputsFromAddressSpent = true;
        }

      }

      if (outputsFromAddressSpent) {
        for (const output of confirmedUnspentOutputs) {
          outputsToConsume[unspentOutput.address][output.id] = output;
        }
      }
    }

    return outputsToConsume;
  }


  public buildOutputs(remainderAddress: string, destinationAddress: string, iotas: bigint, consumedFunds: { [color: string]: bigint; }): {
    [address: string]: {
      /**
       * The color.
       */
      color: string;
      /**
       * The value.
       */
      value: bigint;

      shouldShipPayload: boolean;
    }[];
  } {
    const outputsByColor: { [address: string]: { [color: string]: bigint; }; } = {};

    // build outputs for destinations

    if (!outputsByColor[destinationAddress]) {
      outputsByColor[destinationAddress] = {};
    }


    if (!outputsByColor[destinationAddress][Colors.IOTA_COLOR_STRING]) {
      outputsByColor[destinationAddress][Colors.IOTA_COLOR_STRING] = BigInt(0);
    }
    outputsByColor[destinationAddress][Colors.IOTA_COLOR_STRING] += iotas;

    consumedFunds[Colors.IOTA_COLOR_STRING] -= iotas;
    if (consumedFunds[Colors.IOTA_COLOR_STRING] === BigInt(0)) {
      delete consumedFunds[Colors.IOTA_COLOR_STRING];
    }



    // build outputs for remainder
    if (Object.keys(consumedFunds).length > 0) {
      if (!remainderAddress) {
        throw new Error("No remainder address available");
      }
      if (!outputsByColor[remainderAddress]) {
        outputsByColor[remainderAddress] = {};
      }
      for (const consumed in consumedFunds) {
        if (!outputsByColor[remainderAddress][consumed]) {
          outputsByColor[remainderAddress][consumed] = BigInt(0);
        }
        outputsByColor[remainderAddress][consumed] += consumedFunds[consumed];
      }
    }

    // construct result
    const outputsBySlice: {
      [address: string]: {
        /**
         * The color.
         */
        color: string;
        /**
         * The value.
         */
        value: bigint;

        shouldShipPayload: boolean;
      }[];
    } = {};

    for (const address in outputsByColor) {
      outputsBySlice[address] = [];
      for (const color in outputsByColor[address]) {
        outputsBySlice[address].push({
          color,
          value: outputsByColor[address][color],
          shouldShipPayload: address == destinationAddress
        });
      }
    }

    return outputsBySlice;
  }

  public buildInputs(outputsToUseAsInputs: { [address: string]: { [outputID: string]: IWalletOutput; }; }): {
    /**
     * The inputs to send.
     */
    inputs: string[];
    /**
     * The fund that were consumed.
     */
    consumedFunds: { [color: string]: bigint; };
  } {
    const inputs: string[] = [];
    const consumedFunds: { [color: string]: bigint; } = {};

    for (const address in outputsToUseAsInputs) {
      for (const outputID in outputsToUseAsInputs[address]) {
        inputs.push(outputID);

        for (const balance of outputsToUseAsInputs[address][outputID].balances) {
          if (!consumedFunds[balance.color]) {
            consumedFunds[balance.color] = balance.value;
          } else {
            consumedFunds[balance.color] += balance.value;
          }
        }
      }
    }

    inputs.sort((a, b) => Base58.decode(a).compare(Base58.decode(b)));

    return { inputs, consumedFunds };
  }
}
