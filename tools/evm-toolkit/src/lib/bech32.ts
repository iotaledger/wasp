import {
  SingleNodeClient,
  Bech32Helper,
  type IEd25519Address,
} from '@iota/iota.js';

import { Converter } from '@iota/util.js';

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
