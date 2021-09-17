import type { IPanelDataItem } from "./IPanelDataItem"
import type { IButton } from "./IButton"
import type { IPanel } from "./IPanel"

export interface IValuePanel extends IPanel {
    /**
     * The general data.
     */
    data: IPanelDataItem;

    /**
     * The buttons in the panel.
     */
    buttons?: IButton[];
}
