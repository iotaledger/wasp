import {
    METADATA_FEATURE_TYPE, type IMetadataFeature, type IndexerPluginClient, type INftOutput, type SingleNodeClient
} from '@iota/iota.js';
import { Converter } from '@iota/util.js';

import { NFTMetadataCache } from './constants';
import type { INFTIRC27 } from './interfaces';

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

        tokenData.outputID = nft.items[0];
        return tokenData as INFTIRC27;
    }

    throw new Error('Could not find NFT metadata');
}
