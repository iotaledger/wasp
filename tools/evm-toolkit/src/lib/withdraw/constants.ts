import multiCallAbiAsText from '$lib/assets/multicall3.json?raw';
import iscAbiAsText from '$lib/assets/ISCSandbox.abi?raw';

export const gasFee = 21000;
export const iscAbi = JSON.parse(iscAbiAsText);
export const multiCallAbi = JSON.parse(multiCallAbiAsText).abi;
export const iscContractAddress = '0x1074000000000000000000000000000000000000';
