import type { IPanelDataItem } from "./IPanelDataItem";
import type { ITypeBase } from "./ITypeBase";

/**
 * The global type for the entries panel.
 */
export const WALLET_PANEL_TYPE = 0;

export interface IWalletPanel extends ITypeBase<0> {
    /**
     * The general data.
     */
    data: IPanelDataItem[];
}
