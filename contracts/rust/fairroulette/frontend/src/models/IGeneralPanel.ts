import type { IPanelDataItem } from "./IPanelDataItem"
import type { IPanel } from "./IPanel"

export interface IGeneralPanel extends IPanel {
    /**
     * The general data.
     */
    data: IPanelDataItem[];
}
