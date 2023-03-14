import iscAbiAsText from '../../assets/ISCSandbox.abi?raw';
import multiCallAbiAsText from '../../assets/multicall3.json?raw';

export const gasFee = 21000;
export const iscAbi = JSON.parse(iscAbiAsText);
export const multiCallAbi = JSON.parse(multiCallAbiAsText).abi;
export const iscContractAddress = '0x1074000000000000000000000000000000000000';
