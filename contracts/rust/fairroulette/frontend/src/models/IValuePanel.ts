import type { IPanelDataItem } from "./IPanelDataItem"
import type { IButton } from "./IButton"

export interface IBalancePanel extends IPanel {
    /**
     * The general data.
     */
    data: IPanelDataItem;

    /**
     * The buttons in the panel.
     */
    buttons?: IButton[];
}
