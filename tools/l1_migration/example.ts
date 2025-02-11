import { Transaction } from "@iota/iota-sdk/transactions";

const STARDUST_PACKAGE_ID = '0x107a';
const NFT_OUTPUT_MODULE_NAME = 'nft_output';
const NFT_OUTPUT_STRUCT_NAME = 'NftOutput';

const gasTypeTag = '0x2::iota::IOTA';
const nftOutputStructTag = `${STARDUST_PACKAGE_ID}::${NFT_OUTPUT_MODULE_NAME}::${NFT_OUTPUT_STRUCT_NAME}<${gasTypeTag}>`;
const foundryCapTypeTag = '0x2::coin_manager::CoinManagerTreasuryCap';

const aliasOutputObjectId = "";
const sender = "";
const nftOutputObjectId = "";

// Create the ptb.
    const tx = new Transaction();

    // Extract alias output assets.
    const typeArgs = [gasTypeTag];
    const args = [tx.object(aliasOutputObjectId)]
    const extractedAliasOutputAssets = tx.moveCall({
        target: `${STARDUST_PACKAGE_ID}::alias_output::extract_assets`,
        typeArguments: typeArgs,
        arguments: args,
    });

    // Extract assets.
    const extractedBaseToken = extractedAliasOutputAssets[0];
    const extractedNativeTokensBag = extractedAliasOutputAssets[1];
    const alias = extractedAliasOutputAssets[2];

    // Extract the IOTA balance.
    const iotaCoin = tx.moveCall({
        target: '0x2::coin::from_balance',
        typeArguments: typeArgs,
        arguments: [extractedBaseToken],
    });

    // Transfer the IOTA balance to the sender.
    tx.transferObjects([iotaCoin], tx.pure.address(sender));

    // Cleanup the bag by destroying it.
    tx.moveCall({
        target: '0x2::bag::destroy_empty',
        typeArguments: [],
        arguments: [extractedNativeTokensBag],
    });

    // Unlock the nft output.
    const aliasArg = alias;
    const nftOutputArg = tx.object(nftOutputObjectId);

    const nftOutput = tx.moveCall({
        target: `${STARDUST_PACKAGE_ID}::address_unlock_condition::unlock_alias_address_owned_nft`,
        typeArguments: typeArgs,
        arguments: [aliasArg, nftOutputArg],
    });

    // Transferring alias asset.
    tx.transferObjects([alias], tx.pure.address(sender));

    // Extract the assets from the NftOutput (base token, native tokens bag, nft asset itself).
    const extractedAssets = tx.moveCall({
        target: `${STARDUST_PACKAGE_ID}::nft_output::extract_assets`,
        typeArguments: typeArgs,
        arguments: [nftOutput],
    });

    // If the nft output can be unlocked, the command will be successful and will
    // return a `base_token` (i.e., IOTA) balance and a `Bag` of native tokens and
    // related nft object.

    const extractedBaseToken2 = extractedAssets[0];
    const extractedNativeTokensBag2 = extractedAssets[1];
    const nftAsset = extractedAssets[2];

    // Extract the IOTA balance.
    const iotaCoin2 = tx.moveCall({
        target: '0x2::coin::from_balance',
        typeArguments: typeArgs,
        arguments: [extractedBaseToken2],
    });

    // Transfer the IOTA balance to the sender.
    tx.transferObjects([iotaCoin2], tx.pure.address(sender));

    // Cleanup the bag because it is empty.
    tx.moveCall({
        target: '0x2::bag::destroy_empty',
        typeArguments: [],
        arguments: [extractedNativeTokensBag2],
    });

    // Transferring nft asset.
    tx.transferObjects([nftAsset], tx.pure.address(sender));

