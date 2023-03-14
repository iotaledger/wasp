import type { Contract } from 'web3-eth-contract';
import {
  web3,
} from 'svelte-web3';
import type { Eth } from 'web3-eth';
import { hNameFromString } from '../../../lib/hname';
import { evmAddressToAgentID } from '../../../lib/evm';
import { getBalanceParameters, withdrawParameters } from './../parameters';
import { getNativeTokenMetaData, type INativeToken, type INFT } from '../../../lib/native_token';
import { Converter } from '@iota/util.js';
import { NativeTokenIDLength } from '../../../lib/constants';
import type { IndexerPluginClient, SingleNodeClient } from '@iota/iota.js';
import { gasFee, iscAbi, iscContractAddress } from './../constants';


import {
  Multicall,
  type ContractCallResults,
  type ContractCallContext,
} from 'ethereum-multicall';
import type Web3 from 'web3';

export type NFTDict = [string, string][];

export interface IWithdrawResponse {
  blockHash: string;
  blockNumber: number;
  contractAddress: string;
  cumulativeGasUsed: number;
  from: string;
  gasUsed: number;
  logsBloom: string;
  status: boolean;
  to: string;
  transactionHash: string;
  transactionIndex: number;
  events: unknown;
}

export class ISCMagic {
  private readonly contract: Contract;
  private readonly multicall: Contract;

  constructor(contract: Contract, multicall: Contract) {
    this.contract = contract;
    this.multicall = multicall;
  }

  public async getBaseTokens(eth: Eth, account: string): Promise<number> {
    const addressBalance = await eth.getBalance(account);
    const balance = BigInt(addressBalance) / BigInt(1e12);

    return Number(balance);
  }

  public async getNativeTokens(nodeClient: SingleNodeClient, indexerClient: IndexerPluginClient, account: string) {
    const accountsCoreContract = hNameFromString('accounts');
    const getBalanceFunc = hNameFromString('balance');
    const agentID = evmAddressToAgentID(account);

    const parameters = getBalanceParameters(agentID);

    const nativeTokenResult = await this.contract.methods
      .callView(accountsCoreContract, getBalanceFunc, parameters)
      .call();

    const nativeTokens: INativeToken[] = [];

    for (let item of nativeTokenResult.items) {
      const id = item.key;
      const idBytes = Converter.hexToBytes(id);

      if (idBytes.length != NativeTokenIDLength) {
        continue;
      }

      var nativeToken: INativeToken = {
        // TODO: BigInt is required for native tokens, but it causes problems with the range slider. This needs to be adressed before shipping.
        amount: BigInt(item.value),
        id: id,
        metadata: await getNativeTokenMetaData(nodeClient, indexerClient, id),
      };

      nativeTokens.push(nativeToken);
    }

    return nativeTokens;
  }

  public async getNFTs(account: string) {
    const accountsCoreContract = hNameFromString('accounts');
    const getAccountNFTsFunc = hNameFromString('accountNFTs');
    const agentID = evmAddressToAgentID(account);

    let parameters = getBalanceParameters(agentID);

    const NFTsResult = await this.contract.methods
      .callView(accountsCoreContract, getAccountNFTsFunc, parameters)
      .call();

    const nfts = NFTsResult.items as NFTDict;

    // The 'i' parameter returns the length of the nft id array, but we can just filter that out
    // and go through the list dynamically.
    const nftIds = nfts.filter(x => Converter.hexToUtf8(x[0]) != 'i');
    const availableNFTs = nftIds.map(x => <INFT>{ id: x[1] });

    return availableNFTs;
  }

  public async withdrawMulticall(web3Instance: Web3, multicallAddress: string, nodeClient: SingleNodeClient, receiverAddress: string, baseTokens: number, nativeTokens: INativeToken[], nfts: INFT[]) {
    const sendABI = iscAbi.find(x => x.name == 'send');
    const contractCallContext: ContractCallContext =
    {
      reference: 'send',
      contractAddress: iscContractAddress,

      abi: iscAbi,
      calls: []
    };

    let baseTokensSent = 0;

    for (let nft of nfts) {
      const parameters = await withdrawParameters(
        nodeClient,
        receiverAddress,
        gasFee,
        gasFee * 2,
        [],
        nft,
      );

      baseTokensSent += gasFee * 2;

      contractCallContext.calls.push({
        reference: 'send', methodName: 'send', methodParameters: parameters,
      })
    }

    const lastParameters = await withdrawParameters(
      nodeClient,
      receiverAddress,
      gasFee,
      baseTokens - baseTokensSent,
      nativeTokens,
      null,
    );

    console.log(contractCallContext)

    /*contractCallContext.calls.push({
      reference: 'send', methodName: 'send', methodParameters: lastParameters,
    });*/

    const multicall = new Multicall({ web3Instance: web3Instance, tryAggregate: false, multicallCustomContractAddress: multicallAddress });
    const results: ContractCallResults = await multicall.call(contractCallContext);

    console.log(results);

  }


  public async withdrawMulticallSelfCall(web3Instance: Web3, multicallAddress: string, nodeClient: SingleNodeClient, receiverAddress: string, baseTokens: number, nativeTokens: INativeToken[], nfts: INFT[]) {
    const sendABI = iscAbi.find(x => x.name == 'send');
    const contractCallContext: ContractCallContext =
    {
      reference: 'send',
      contractAddress: iscContractAddress,

      abi: iscAbi,
      calls: []
    };

    let baseTokensSent = 0;

    const lastParameters = await withdrawParameters(
      nodeClient,
      receiverAddress,
      gasFee,
      baseTokens - baseTokensSent,
      nativeTokens,
      null,
    );

    contractCallContext.calls.push({
      reference: 'send', methodName: 'send', methodParameters: lastParameters,
    });

    const multicall = new Multicall({ web3Instance: web3Instance, tryAggregate: false, multicallCustomContractAddress: multicallAddress });
    const results: ContractCallResults = await multicall.call(contractCallContext);

    console.log(results);

  }

  public async withdraw(nodeClient: SingleNodeClient, receiverAddress: string, baseTokens: number, nativeTokens: INativeToken[], nft?: INFT) {
    const parameters = await withdrawParameters(
      nodeClient,
      receiverAddress,
      gasFee,
      baseTokens,
      nativeTokens,
      nft,
    );

    let result = await this.contract.methods.send(...parameters).send();

    return result as IWithdrawResponse;
  }
}