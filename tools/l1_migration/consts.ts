export const STARDUST_PACKAGE_ID = '0x107a';
export const NFT_OUTPUT_MODULE_NAME = 'nft_output';
export const NFT_OUTPUT_STRUCT_NAME = 'NftOutput';

export const gasTypeTag = '0x2::iota::IOTA';
export const nftOutputStructTag = `${STARDUST_PACKAGE_ID}::${NFT_OUTPUT_MODULE_NAME}::${NFT_OUTPUT_STRUCT_NAME}<${gasTypeTag}>`;
export const foundryCapTypeTag = '0x2::coin_manager::CoinManagerTreasuryCap';