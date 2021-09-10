import type { IWalletOutputBalance } from "./IWalletOutputBalance";
import type { IWalletOutputInclusionState } from "./IWalletOutputInclusionState";

export interface IWalletOutput {
    /**
     * The id.
     */
    id: string;

    /**
     * The balances.
     */
    balances: IWalletOutputBalance[];

    /**
     * Inclusion state.
     */
    inclusionState: IWalletOutputInclusionState;
}
