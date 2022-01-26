// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

import {BytesDecoder, BytesEncoder} from "./bytes";
import {Convert} from "./convert";
import {ScFuncCallContext, ScViewCallContext} from "./contract";
import {ScAddress, ScAgentID, ScChainID, ScColor, ScHash, ScHname, ScRequestID} from "./hashtypes";
import {log, OBJ_ID_ROOT, OBJ_ID_STATE, panic} from "./host";
import {ScImmutableColorArray, ScImmutableMap} from "./immutable";
import * as keys from "./keys";
import {ScMutableMap} from "./mutable";

// all access to the objects in host's object tree starts here
export let ROOT = new ScMutableMap(OBJ_ID_ROOT);

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// retrieves any information that is related to colored token balances
export class ScBalances {
    balances: ScImmutableMap;

    constructor(id: keys.Key32) {
        this.balances = ROOT.getMap(id).immutable()
    }

    // retrieve the balance for the specified token color
    balance(color: ScColor): u64 {
        return this.balances.getUint64(color).value();
    }

    // retrieve an array of all token colors that have a non-zero balance
    colors(): ScImmutableColorArray {
        return this.balances.getColorArray(keys.KEY_COLOR);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// passes token transfer information to a function call
export class ScTransfers {
    transfers: ScMutableMap;

    // create a new transfers object ready to add token transfers
    constructor() {
        this.transfers = ScMutableMap.create();
    }

    // create a new transfers object from a balances object
    static fromBalances(balances: ScBalances): ScTransfers {
        let transfers = new ScTransfers();
        let colors = balances.colors();
        for (let i:u32 = 0; i < colors.length(); i++) {
            let color = colors.getColor(i).value();
            transfers.set(color, balances.balance(color));
        }
        return transfers;
    }

    // create a new transfers object and initialize it with the specified amount of iotas
    static iotas(amount: u64): ScTransfers {
        return ScTransfers.transfer(ScColor.IOTA, amount);
    }

    // create a new transfers object and initialize it with the specified token transfer
    static transfer(color: ScColor, amount: u64): ScTransfers {
        let transfer = new ScTransfers();
        transfer.set(color, amount);
        return transfer;
    }

    // set the specified colored token transfer in the transfers object
    // note that this will overwrite any previous amount for the specified color
    set(color: ScColor, amount: u64): void {
        this.transfers.getUint64(color).setValue(amount);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// provides access to utility functions that are handled by the host
export class ScUtility {
    utility: ScMutableMap;

    constructor() {
        this.utility = ROOT.getMap(keys.KEY_UTILITY)
    }

    // decodes the specified base58-encoded string value to its original bytes
    base58Decode(value: string): u8[] {
        return this.utility.callFunc(keys.KEY_BASE58_DECODE, Convert.fromString(value));
    }

    // encodes the specified bytes to a base-58-encoded string
    base58Encode(value: u8[]): string {
        let result = this.utility.callFunc(keys.KEY_BASE58_ENCODE, value);
        return Convert.toString(result);
    }

    // retrieves the address for the specified BLS public key
    blsAddressFromPubkey(pubKey: u8[]): ScAddress {
        let result = this.utility.callFunc(keys.KEY_BLS_ADDRESS, pubKey);
        return ScAddress.fromBytes(result);
    }

    // aggregates the specified multiple BLS signatures and public keys into a single one
    blsAggregateSignatures(pubKeysBin: u8[][], sigsBin: u8[][]): u8[][] {
        let encode = new BytesEncoder();
        encode.int32(pubKeysBin.length);
        for (let i = 0; i < pubKeysBin.length; i++) {
            encode.bytes(pubKeysBin[i]);
        }
        encode.int32(sigsBin.length as i32);
        for (let i = 0; i < sigsBin.length; i++) {
            encode.bytes(sigsBin[i]);
        }
        let result = this.utility.callFunc(keys.KEY_BLS_AGGREGATE, encode.data());
        let decode = new BytesDecoder(result);
        return [decode.bytes(), decode.bytes()];
    }

    // checks if the specified BLS signature is valid
    blsValidSignature(data: u8[], pubKey: u8[], signature: u8[]): boolean {
        let encode = new BytesEncoder();
        encode.bytes(data);
        encode.bytes(pubKey);
        encode.bytes(signature);
        let result = this.utility.callFunc(keys.KEY_BLS_VALID, encode.data());
        return (result[0] & 0x01) != 0;
    }

    // retrieves the address for the specified ED25519 public key
    ed25519AddressFromPubkey(pubKey: u8[]): ScAddress {
        let result = this.utility.callFunc(keys.KEY_ED25519_ADDRESS, pubKey);
        return ScAddress.fromBytes(result);
    }

    // checks if the specified ED25519 signature is valid
    ed25519ValidSignature(data: u8[], pubKey: u8[], signature: u8[]): boolean {
        let encode = new BytesEncoder();
        encode.bytes(data);
        encode.bytes(pubKey);
        encode.bytes(signature);
        let result = this.utility.callFunc(keys.KEY_ED25519_VALID, encode.data());
        return (result[0] & 0x01) != 0;
    }

    // hashes the specified value bytes using BLAKE2b hashing and returns the resulting 32-byte hash
    hashBlake2b(value: u8[]): ScHash {
        let hash = this.utility.callFunc(keys.KEY_HASH_BLAKE2B, value);
        return ScHash.fromBytes(hash);
    }

    // hashes the specified value bytes using SHA3 hashing and returns the resulting 32-byte hash
    hashSha3(value: u8[]): ScHash {
        let hash = this.utility.callFunc(keys.KEY_HASH_SHA3, value);
        return ScHash.fromBytes(hash);
    }

    // calculates 32-bit hash for the specified name string
    hname(name: string): ScHname {
        let result = this.utility.callFunc(keys.KEY_HNAME, Convert.fromString(name));
        return ScHname.fromBytes(result);
    }
}

// wrapper function for simplified internal access to base58 encoding
export function base58Encode(bytes: u8[]): string {
    return new ScFuncContext().utility().base58Encode(bytes);
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// shared interface part of ScFuncContext and ScViewContext
export class ScBaseContext {
    // retrieve the agent id of this contract account
    accountID(): ScAgentID {
        return ROOT.getAgentID(keys.KEY_ACCOUNT_ID).value();
    }

    // access the current balances for all token colors
    balances(): ScBalances {
        return new ScBalances(keys.KEY_BALANCES);
    }

    // retrieve the chain id of the chain this contract lives on
    chainID(): ScChainID {
        return ROOT.getChainID(keys.KEY_CHAIN_ID).value();
    }

    // retrieve the agent id of the owner of the chain this contract lives on
    chainOwnerID(): ScAgentID {
        return ROOT.getAgentID(keys.KEY_CHAIN_OWNER_ID).value();
    }

    // retrieve the hname of this contract
    contract(): ScHname {
        return ROOT.getHname(keys.KEY_CONTRACT).value();
    }

    // retrieve the agent id of the creator of this contract
    contractCreator(): ScAgentID {
        return ROOT.getAgentID(keys.KEY_CONTRACT_CREATOR).value();
    }

    // logs informational text message in the log on the host
    log(text: string): void {
        log(text);
    }

    // logs error text message in the log on the host and then panics
    panic(text: string): void {
        panic(text);
    }

    // retrieve parameters that were passed to the smart contract function
    params(): ScImmutableMap {
        return ROOT.getMap(keys.KEY_PARAMS).immutable();
    }

    // panics with specified message if specified condition is not satisfied
    require(cond: boolean, msg: string): void {
        if (!cond) {
            panic(msg);
        }
    }

    // map that holds any results returned by the smart contract function
    results(): ScMutableMap {
        return ROOT.getMap(keys.KEY_RESULTS);
    }

    // deterministic time stamp fixed at the moment of calling the smart contract
    timestamp(): i64 {
        return ROOT.getInt64(keys.KEY_TIMESTAMP).value();
    }

    // logs debugging trace text message in the log on the host
    // similar to log() except this will only show in the log in a special debug mode
    trace(text: string): void {
        trace(text);
    }

    // access diverse utility functions provided by the host
    utility(): ScUtility {
        return new ScUtility();
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract interface with mutable access to state
export class ScFuncContext extends ScBaseContext implements ScViewCallContext, ScFuncCallContext {
    canCallFunc(): void {
        panic!("canCallFunc");
    }

    canCallView(): void {
        panic!("canCallView");
    }

    // synchronously calls the specified smart contract function,
    // passing the provided parameters and token transfers to it
    call(hcontract: ScHname, hfunction: ScHname, params: ScMutableMap | null, transfer: ScTransfers | null): ScImmutableMap {
        let encode = new BytesEncoder();
        encode.hname(hcontract);
        encode.hname(hfunction);
        encode.int32((params === null) ? 0 : params.mapID());
        encode.int32((transfer === null) ? 0 : transfer.transfers.mapID());
        ROOT.getBytes(keys.KEY_CALL).setValue(encode.data());
        return ROOT.getMap(keys.KEY_RETURN).immutable();
    }

    // retrieve the agent id of the caller of the smart contract
    caller(): ScAgentID {
        return ROOT.getAgentID(keys.KEY_CALLER).value();
    }

    // deploys a new instance of the specified smart contract on the current chain
    // the provided parameters are passed to the smart contract "init" function
    deploy(programHash: ScHash, name: string, description: string, params: ScMutableMap | null): void {
        let encode = new BytesEncoder();
        encode.hash(programHash);
        encode.string(name);
        encode.string(description);
        encode.int32((params === null) ? 0 : params.mapID());
        ROOT.getBytes(keys.KEY_DEPLOY).setValue(encode.data());
    }

    // signals an event on the host that external entities can subscribe to
    event(text: string): void {
        ROOT.getString(keys.KEY_EVENT).setValue(text);
    }

    // access the incoming balances for all token colors
    incoming(): ScBalances {
        return new ScBalances(keys.KEY_INCOMING);
    }

    // retrieve the tokens that were minted in this transaction
    minted(): ScBalances {
        return new ScBalances(keys.KEY_MINTED);
    }

    // asynchronously calls the specified smart contract function,
    // passing the provided parameters and token transfers to it
    // it is possible to schedule the call for a later execution by specifying a delay
    post(chainID: ScChainID, hcontract: ScHname, hfunction: ScHname, params: ScMutableMap | null, transfer: ScTransfers, delay: u32): void {
        let encode = new BytesEncoder();
        encode.chainID(chainID);
        encode.hname(hcontract);
        encode.hname(hfunction);
        encode.int32((params === null) ? 0 : params.mapID());
        encode.int32(transfer.transfers.mapID());
        encode.uint32(delay);
        ROOT.getBytes(keys.KEY_POST).setValue(encode.data());
    }

    // generates a random value from 0 to max (exclusive max) using a deterministic RNG
    random(max: u64): u64 {
        if (max == 0) {
            this.panic("random: max parameter should be non-zero");
        }
        let state = new ScMutableMap(OBJ_ID_STATE);
        let rnd = state.getBytes(keys.KEY_RANDOM);
        let seed = rnd.value();
        if (seed.length == 0) {
            seed = ROOT.getBytes(keys.KEY_RANDOM).value();
        }
        rnd.setValue(this.utility().hashSha3(seed).toBytes());
        return Convert.toI64(seed.slice(0, 8)) as u64 % max
    }

    // retrieve the request id of this transaction
    requestID(): ScRequestID {
        return ROOT.getRequestID(keys.KEY_REQUEST_ID).value();
    }

    // access mutable state storage on the host
    state(): ScMutableMap {
        return ROOT.getMap(keys.KEY_STATE);
    }

    // transfers the specified tokens to the specified Tangle ledger address
    transferToAddress(address: ScAddress, transfer: ScTransfers): void {
        let transfers = ROOT.getMapArray(keys.KEY_TRANSFERS);
        let tx = transfers.getMap(transfers.length());
        tx.getAddress(keys.KEY_ADDRESS).setValue(address);
        tx.getInt32(keys.KEY_BALANCES).setValue(transfer.transfers.mapID());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
export class ScViewContext extends ScBaseContext implements ScViewCallContext {
    canCallView(): void {
        panic!("canCallView");
    }

    // synchronously calls the specified smart contract view,
    // passing the provided parameters to it
    call(hcontract: ScHname, hfunction: ScHname, params: ScMutableMap | null): ScImmutableMap {
        let encode = new BytesEncoder();
        encode.hname(hcontract);
        encode.hname(hfunction);
        encode.int32((params === null) ? 0 : params.mapID());
        encode.int32(0);
        ROOT.getBytes(keys.KEY_CALL).setValue(encode.data());
        return ROOT.getMap(keys.KEY_RETURN).immutable();
    }

    // access immutable state storage on the host
    state(): ScImmutableMap {
        return ROOT.getMap(keys.KEY_STATE).immutable();
    }
}
