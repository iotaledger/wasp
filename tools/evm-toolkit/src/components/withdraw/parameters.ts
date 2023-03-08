import { Converter } from "@iota/util.js";
import { waspAddrBinaryFromBech32 } from "./../../lib/bech32";
import type { SingleNodeClient } from "@iota/iota.js";
import type { INativeToken } from "../../lib/native_token";

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

export interface INativeTokenWithdraw {
  /**
   * Identifier of the native token.
   */
  ID: string;
  /**
   * Amount of native tokens of the given Token ID.
   */
  amount: bigint;
}

export async function withdrawParameters(nodeClient: SingleNodeClient, receiverAddressBech32: string, gasFee: number, baseTokensToWithdraw: number, nativeTokens: INativeToken[]) {
  const binaryAddress = await waspAddrBinaryFromBech32(nodeClient, receiverAddressBech32);

  /*
    NativeToken[]:
      ID: Tuple[tokenID(string)] (Not just `ID: tokenID`)
      amount: uint256
  */
  const nativeTokenTuple = nativeTokens.map((x) => ({ ID: [x.id], amount: x.amount }));

  const parameters = [
    {
      // Receiver
      data: binaryAddress,
    },
    {
      // Fungible Tokens
      baseTokens: baseTokensToWithdraw - gasFee,
      nativeTokens: nativeTokenTuple,
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