export interface INFTIRC27 {
    standard: string;
    type: string;
    version: string;
    uri: string;
    name: string;

    outputID: string;
}

export interface INFT {
    /**
     * Identifier of the NFT
     */
    id: string;

    outputID: string;

    /**
     * Metadata of the NFT
     */
    metadata?: INFTIRC27;
}

export interface INFTMetaDataCacheMap {
    [Key: string]: INFTIRC27;
}
