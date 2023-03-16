import {
  METADATA_FEATURE_TYPE,
  type IFoundryOutput,
  type IndexerPluginClient,
  type SingleNodeClient,
  type IMetadataFeature,
  type HexEncodedAmount,
  type INftOutput,
} from '@iota/iota.js';
import { Converter } from '@iota/util.js';

export interface INFTIRC27 {
  standard: string;
  type: string;
  version: string;
  uri: string;
  name: string;
}

export interface INFT {
  /**
   * Identifier of the NFT
   */
  id: string;

  /**
   * Metadata of the NFT
   */
  metadata?: INFTIRC27;
}

interface INFTMetaDataCacheMap {
  [Key: string]: INFTIRC27;
}
const NFTMetadataCache: INFTMetaDataCacheMap = {};

export async function getNFTMetadata(client: SingleNodeClient,
  indexer: IndexerPluginClient,
  nftID: string) {
  const nft = await indexer.nft(nftID);

  if (nft.items.length == 0) {
    throw new Error('No outputs found');
  }

  const output = await client.output(nft.items[0]);
  const nftOutput = output.output as INftOutput;

  const metadataFeature = nftOutput.immutableFeatures.find(
    x => x.type == METADATA_FEATURE_TYPE,
  );

  if (metadataFeature) {
    const metaData = (metadataFeature as IMetadataFeature).data;
    const tokenData = JSON.parse(Converter.hexToUtf8(metaData));

    NFTMetadataCache[nftID] = tokenData;

    return tokenData as INFTIRC27;
  }

  throw new Error('Could not find NFT metadata');
}
