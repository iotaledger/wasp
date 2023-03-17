import {
  METADATA_FEATURE_TYPE,
  type IFoundryOutput,
  type IndexerPluginClient,
  type SingleNodeClient,
  type IMetadataFeature,
  type HexEncodedAmount,
} from '@iota/iota.js';
import { Converter } from '@iota/util.js';

export interface INativeTokenIRC30 {
  decimals: number;
  description?: string;
  logo?: string;
  logoUrl?: string;
  name: string;
  standard: string;
  symbol: string;
  url?: string;
}

export interface INativeToken {
  /**
   * Identifier of the native token.
   */
  id: string;
  /**
   * Amount of native tokens of the given Token ID.
   */
  amount: bigint;
  /**
   * Native Token metadata according to IRC30
   */
  metadata?: INativeTokenIRC30;
}

interface INativeTokenMetaDataCacheMap {
  [Key: string]: INativeTokenIRC30;
}
const nativeTokenMetadataCache: INativeTokenMetaDataCacheMap = {};

export async function getNativeTokenMetaData(
  client: SingleNodeClient,
  indexer: IndexerPluginClient,
  nativeTokenID: string,
): Promise<INativeTokenIRC30> {
  if (nativeTokenMetadataCache[nativeTokenID]) {
    return nativeTokenMetadataCache[nativeTokenID];
  }

  const foundryOutputIDs = await indexer.foundry(nativeTokenID);

  if (foundryOutputIDs.items.length == 0) {
    throw new Error('No outputs found');
  }

  const outputID = foundryOutputIDs.items[0];

  const output = await client.output(outputID);
  const foundryOutput = output.output as IFoundryOutput;

  const metadataFeature = foundryOutput.immutableFeatures.find(
    x => x.type == METADATA_FEATURE_TYPE,
  );

  if (metadataFeature) {
    const metaData = (metadataFeature as IMetadataFeature).data;
    const tokenData = JSON.parse(Converter.hexToUtf8(metaData));

    nativeTokenMetadataCache[nativeTokenID] = tokenData;

    return tokenData as INativeTokenIRC30;
  }

  throw new Error('Could not find native token metadata');
}
