import { Buffer } from "../buffer";

import { IFaucetRequest, IFaucetRequestContext, IFaucetResponse } from "./faucet/faucet_models";
import { FaucetHelper } from "./faucet/faucet_helper";

import { IUnspentOutputsRequest, IUnspentOutputsResponse } from "./models/unspent_outputs";
import { IAllowedManaPledgeResponse } from "./models/mana";

import { PoWWorkerManager } from "./pow_web_worker/pow_worker_manager";

import * as requestSender from "../api_common/request_sender";

import { Base58, getAddress, IKeyPair } from "../crypto";

import { IOnLedger, OnLedgerHelper } from "./models/on_ledger";
import { ISendTransactionRequest, ISendTransactionResponse, ITransaction, Transaction } from "./models/transaction";
import { Wallet } from "./wallet/wallet";
import { Colors } from "../colors";
import { AgentID, Configuration, Transfer } from "..";
import { CoreAccountsService } from "../coreaccounts/service";

interface GoShimmerClientConfiguration {
    APIUrl: string;
    SeedUnsafe: Buffer | null;
}

export class GoShimmerClient {
    private coreAccountsService: CoreAccountsService;
    private readonly goShimmerConfiguration: GoShimmerClientConfiguration;
    private readonly powManager: PoWWorkerManager = new PoWWorkerManager();

    constructor(configuration : Configuration, coreAccountsService: CoreAccountsService) {
        this.coreAccountsService = coreAccountsService;
        this.goShimmerConfiguration = { APIUrl: configuration.goShimmerApiUrl, SeedUnsafe: configuration.seed };
    }

    public async getIOTABalance(address: string): Promise<bigint> {
        const iotaBalance = await this.getBalance(address, Colors.IOTA_COLOR_STRING);
        return iotaBalance;
    }

    private async getBalance(address: string, color: string): Promise<bigint> {
        if (color == Colors.IOTA_COLOR) {
            color = Colors.IOTA_COLOR_STRING;
        }

        const unspents = await this.unspentOutputs({ addresses: [address] });
        const currentUnspent = unspents.unspentOutputs.find((x) => x.address.base58 == address);

        const balance = currentUnspent!.outputs
            .filter(
                (o) =>
                    ["ExtendedLockedOutputType", "SigLockedColoredOutputType"].includes(o.output.type) &&
                    typeof o.output.output.balances[color] != "undefined"
            )
            .map((uid) => uid.output.output.balances)
            .reduce((balance: bigint, output) => (balance += BigInt(output[color])), BigInt(0));

        return balance;
    }

    public async unspentOutputs(request: IUnspentOutputsRequest): Promise<IUnspentOutputsResponse> {
        return requestSender.sendRequest<IUnspentOutputsRequest, IUnspentOutputsResponse>(
            this.goShimmerConfiguration.APIUrl,
            "post",
            "ledgerstate/addresses/unspentOutputs",
            request
        );
    }

    public async requestFunds(address: string): Promise<boolean> {
        try {
            const faucetRequestContext = await this.getFaucetRequest(address);
            const response = await this.sendFaucetRequest(faucetRequestContext.faucetRequest);
            const success = response.error === undefined && response.id !== undefined;
            return success;
        } catch (ex: unknown) {
            const error = ex as Error;
            console.error(error.message);
            return false;
        }
    }

    private async getFaucetRequest(address: string): Promise<IFaucetRequestContext> {
        const manaPledge = await this.getAllowedManaPledge();

        const allowedManagePledge = manaPledge.accessMana?.allowed ? manaPledge.accessMana.allowed[0] : "";
        const consenseusManaPledge = manaPledge.consensusMana?.allowed ? manaPledge.consensusMana?.allowed[0] : "";

        const body: IFaucetRequest = {
            accessManaPledgeID: allowedManagePledge,
            consensusManaPledgeID: consenseusManaPledge,
            address: address,
            nonce: -1,
        };

        const poWBuffer = FaucetHelper.ToBuffer(body);

        body.nonce = await this.powManager.requestProofOfWork(12, poWBuffer);

        const result: IFaucetRequestContext = {
            poWBuffer: poWBuffer,
            faucetRequest: body,
        };

        return result;
    }

    private async getAllowedManaPledge(): Promise<IAllowedManaPledgeResponse> {
        return requestSender.sendRequest<null, IAllowedManaPledgeResponse>(this.goShimmerConfiguration.APIUrl, "get", "mana/allowedManaPledge");
    }

    private async sendFaucetRequest(faucetRequest: IFaucetRequest): Promise<IFaucetResponse> {
        const response = await requestSender.sendRequest<IFaucetRequest, IFaucetResponse>(
            this.goShimmerConfiguration.APIUrl,
            "post",
            "faucet",
            faucetRequest
        );
        return response;
    }

    public async postOnLedgerRequest(
        chainId: string,
        payload: IOnLedger,
        transfer: bigint = 1n,
        keyPair: IKeyPair
    ): Promise<string> {
        if (transfer <= 0) {
            transfer = 1n;
        }

        const wallet = new Wallet(this);

        const address = getAddress(keyPair);
        const unspents = await wallet.getUnspentOutputs(address);
        const consumedOutputs = wallet.determineOutputsToConsume(unspents, transfer);
        const { inputs, consumedFunds } = wallet.buildInputs(consumedOutputs);
        const outputs = wallet.buildOutputs(address, chainId, transfer, consumedFunds);

        const tx: ITransaction = {
            version: 0,
            timestamp: BigInt(Date.now()) * 1000000n,
            aManaPledge: Base58.encode(Buffer.alloc(32)),
            cManaPledge: Base58.encode(Buffer.alloc(32)),
            inputs: inputs,
            outputs: outputs,
            chainId: chainId,
            payload: OnLedgerHelper.ToBuffer(payload),
            unlockBlocks: [],
        };

        tx.unlockBlocks = wallet.unlockBlocks(tx, keyPair, address, consumedOutputs, inputs);

        const result = Transaction.bytes(tx);

        const response = await this.sendTransaction({
            txn_bytes: result.toString("base64"),
        });

        if (!response || response.error != undefined) throw Error("Transaction error: " + response.error);
        return response?.transaction_id ?? "";
    }

    private async sendTransaction(request: ISendTransactionRequest): Promise<ISendTransactionResponse> {
        return requestSender.sendRequest<ISendTransactionRequest, ISendTransactionResponse>(
            this.goShimmerConfiguration.APIUrl,
            "post",
            "ledgerstate/transactions",
            request
        );
    }

    public async depositIOTAToAccountInChain(keypair: IKeyPair, destinationAgentID: AgentID, amount: bigint): Promise<boolean> {
        const depositfunc = this.coreAccountsService.deposit();
        depositfunc.agentID(destinationAgentID);
        depositfunc.onLedgerRequest(true);
        depositfunc.transfer(Transfer.iotas(amount));
        depositfunc.sign(keypair);
        const depositRequestID = await depositfunc.post();
        const success = depositRequestID.length > 0;
        return success;
    }
}
