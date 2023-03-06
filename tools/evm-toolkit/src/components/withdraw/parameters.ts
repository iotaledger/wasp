import { Converter } from "@iota/util.js";
import { waspAddrBinaryFromBech32 } from "../../lib/bech32";
import type { SingleNodeClient } from "@iota/iota.js";

export function getBalanceParameters(agentID: Uint8Array) {
  return {
    items: [
      {
        key: Converter.utf8ToBytes("a"),
        value: agentID,
      }
    ],
  }
}

export async function withdrawParameters(nodeClient: SingleNodeClient, receiverAddressBech32: string, baseTokensToWithdraw: number, gasFee: number) {
  const binaryAddress = await waspAddrBinaryFromBech32(nodeClient, receiverAddressBech32);

  let parameters = [
    {
      // Receiver
      data: binaryAddress,
    },
    {
      // Fungible Tokens
      baseTokens: baseTokensToWithdraw - gasFee,
      nativeTokens: [],
      nfts: [],
    },
    false,
    {
      // Metadata
      targetContract: 0,
      entrypoint: 0,
      gasBudget: 0,
      params: {
        items: [],
      },
      allowance: {
        baseTokens: 0,
        nativeTokens: [],
        nfts: [],
      },
    },
    {
      // Options
      timelock: 0,
      expiration: {
        time: 0,
        returnAddress: {
          data: [],
        },
      },
    },
  ];

  return parameters;
}