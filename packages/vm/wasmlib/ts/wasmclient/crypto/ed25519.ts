import nacl from "tweetnacl";
import { AgentID } from "..";
import { Buffer } from "../buffer";
import { Base58 } from "./base58";
import { Hash } from "./hash";
import { Seed } from "./seed";

export interface IKeyPair {
  publicKey: Buffer;
  secretKey: Buffer;
}

/**
 * Calculates the AgentID for the key pair's address.
 * @param keyPair The key pair used to get the address and calculate the AgentID.
 * @returns AgentID.
 */
 export function getAgentId(keyPair: IKeyPair) : AgentID {
  const address = getAddress(keyPair);
  const addressBuffer = Base58.decode(address);
  const hNameBuffer = Buffer.alloc(4);
  const agentIdBuffer = Buffer.concat([addressBuffer,hNameBuffer]);
  const agentId = Base58.encode(agentIdBuffer);
  return agentId;
}

/**
 * Calculates the address for key pair
 * @param keyPair The key pair used to calculate the address
 * @returns The generated address.
 */
 export function getAddress(keyPair: IKeyPair): string {
  const publicKeyBuffer = Buffer.from(keyPair.publicKey);
  return getAddressFromPublicKeyBuffer(publicKeyBuffer);
}

export function getAddressFromPublicKeyBuffer(publicKeyBuffer: Buffer): string {
  const digest = Hash.from(publicKeyBuffer);

  const buffer = Buffer.alloc(Seed.SEED_SIZE + 1);
  buffer[0] = ED25519.VERSION;
  Buffer.from(digest).copy(buffer, 1);

  return Base58.encode(buffer);
}

/**
 * Class to help with ED25519 Signature scheme.
 */
export class ED25519 {
  public static VERSION: number = 0;
  public static PUBLIC_KEY_SIZE: number = 32;
  public static SIGNATURE_SIZE: number = 64;

  /**
   * Generate a key pair from the seed.
   * @param seed The seed to generate the key pair from.
   * @returns The key pair.
   */
  public static keyPairFromSeed(seed: Buffer): IKeyPair {
    const signKeyPair = nacl.sign.keyPair.fromSeed(seed);

    return {
      publicKey: Buffer.from(signKeyPair.publicKey),
      secretKey: Buffer.from(signKeyPair.secretKey),
    };
  }

  /**
   * Privately sign the data.
   * @param keyPair The key pair to sign with.
   * @param buffer The data to sign.
   * @returns The signature.
   */
  public static privateSign(keyPair: IKeyPair, buffer: Buffer): Buffer {
    return Buffer.from(nacl.sign.detached(buffer, keyPair.secretKey));
  }
}
