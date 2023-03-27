import type { SingleNodeClient } from '@iota/iota.js';
import {
  Bech32Helper,
  type IEd25519Address
} from '@iota/iota.js';
import { Converter } from '@iota/util.js';

import type { INativeToken } from '$lib/native-token';

export function getBalanceParameters(agentID: Uint8Array) {
  return {
    items: [
      {
        key: Converter.utf8ToBytes('a'),
        value: agentID,
      },
    ],
  };
}

export async function withdrawParameters(
  nodeClient: SingleNodeClient,
  receiverAddressBech32: string,
  gasFee: number,
  baseTokensToWithdraw: number,
  nativeTokens: INativeToken[],
  nftID?: string,
) {
  const binaryAddress = await waspAddrBinaryFromBech32(
    nodeClient,
    receiverAddressBech32,
  );

  /*
    NativeToken[]:
      ID: Tuple[tokenID(string)] (Not just `ID: tokenID`)
      amount: uint256
  */
  const nativeTokenTuple = nativeTokens.map(x => ({
    ID: [x.id],
    amount: x.amount,
  }));

  const nftIDParam = [];
  if (nftID) {
    nftIDParam.push(nftID);
  }

  const parameters = [
    {
      // Receiver
      data: binaryAddress,
    },
    {
      // Fungible Tokens
      baseTokens: baseTokensToWithdraw - gasFee,
      nativeTokens: nativeTokenTuple,
      nfts: nftIDParam,
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

export async function waspAddrBinaryFromBech32(
  nodeClient: SingleNodeClient,
  bech32String: string,
) {
  const protocolInfo = await nodeClient.info();

  const receiverAddr = Bech32Helper.addressFromBech32(
    bech32String,
    protocolInfo.protocol.bech32Hrp,
  );

  const address: IEd25519Address = receiverAddr as IEd25519Address;

  const receiverAddrBinary = Converter.hexToBytes(address.pubKeyHash);
  //  // AddressEd25519 denotes an Ed25519 address.
  // AddressEd25519 AddressType = 0
  // // AddressAlias denotes an Alias address.
  // AddressAlias AddressType = 8
  // // AddressNFT denotes an NFT address.
  // AddressNFT AddressType = 16
  //
  // 0 is the ed25519 prefix
  return new Uint8Array([0, ...receiverAddrBinary]);
}
