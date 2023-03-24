import {
    METADATA_FEATURE_TYPE,
    type IFoundryOutput, type IMetadataFeature, type IndexerPluginClient,
    type SingleNodeClient
} from '@iota/iota.js';
import { Converter } from '@iota/util.js';

import { nativeTokenMetadataCache } from './constants';
import type { INativeTokenIRC30 } from './interfaces';

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