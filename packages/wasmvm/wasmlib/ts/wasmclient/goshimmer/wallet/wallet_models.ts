import {ColorCollection} from "../../colors";

export interface IWalletOutputInclusionState {
    solid?: boolean;
    confirmed?: boolean;
    rejected?: boolean;
    liked?: boolean;
    conflicting?: boolean;
    finalized?: boolean;
    preferred?: boolean;
}

export interface IWalletOutput {
    /**
     * The id.
     */
    id: string;

    /**
     * The balances.
     */
    balances: ColorCollection;

    /**
     * Inclusion state.
     */
    inclusionState: IWalletOutputInclusionState;
}

export interface IWalletAddressOutput {
    /**
     * The address.
     */
    address: string;

    /**
     * The outputs.
     */
    outputs: IWalletOutput[];
}
