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

export interface INativeTokenMetaDataCacheMap {
    [Key: string]: INativeTokenIRC30;
}
