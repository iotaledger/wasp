import { Converter } from '@iota/util.js';

export function evmAddressToAgentID(evmStoreAccount: string): Uint8Array {
  // This function constructs an AgentID that is required to be used with contracts
  // Wasp understands different AgentID types and each AgentID needs to provide a certain ID that describes it's address type.
  // In the case of EVM addresses it's ID 3.
  const agentIDKindEthereumAddress = 3;

  const receiverAddrBinary = Converter.hexToBytes(evmStoreAccount);
  const addressBytes = new Uint8Array([
    agentIDKindEthereumAddress,
    ...receiverAddrBinary,
  ]);

  return addressBytes;
}
