import type { IButton } from "./IButton";
import type { IPanelDataItem } from "./IPanelDataItem";
import type { ITypeBase } from "./ITypeBase";

/**
 * The global type for the entries panel.
 */
export const BALANCE_PANEL_TYPE = 1;
export interface IBalancePanel extends ITypeBase<1> {
    /**
     * The general data.
     */
    data: IPanelDataItem;

    /**
     * The buttons in the panel.
     */
    buttons?: IButton[];
}
