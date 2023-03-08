import type { SetIntervalAsyncTimer } from "set-interval-async";
import type { INativeToken } from "../../lib/native_token";
import type { Contract } from 'web3-eth-contract';

export interface WithdrawState {
  /**
   * The current available base token balance of the user.
   */
  availableBaseTokens: number;

  /**
   * The current available native tokens and balance of the user.
   */
  availableNativeTokens: INativeToken[];

  /**
   * The reference to the ISC magic contract used for contract invocations.
   */
  contract: Contract;

  /**
   * The EVM chain ID.
   */
  evmChainID: number;

  /**
   * Whether or not Metamask is connected to the page.
   */
  isMetamaskConnected: boolean;

  /**
   * Whether or not the page is loading (Getting initial balance, connecting to wallet, initializing contract, ..)
   */
  isLoading: boolean;

  /**
   * The handle of the async balance polling interval.
   */
  balancePollingHandle: SetIntervalAsyncTimer<[]>;
}

export interface WithdrawFormInput {
  /**
   * [Form] The address to send funds to
   */
  receiverAddress: string;

  /**
   * [Form] The amount of base tokens to send.
   */
  baseTokensToSend: number;

  /**
   * [Form] A map of native tokens to send.
   */
  nativeTokensToSend: { [key: string]: number };
}