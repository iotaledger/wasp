import type { IPanelDataItem } from "./IPanelDataItem"
import type { IPanel } from "./IPanel"

export interface IDetailsPanel extends IPanel {
    /**
     * The title of the panel.
     */
     title?: string;

     /**
      * Is items ordered
      */
     ordered?: boolean;
    /**
     * The general data.
     */
    data: (IPanelDataItem & {
        /**
         * The tag that belongs to the item.
         */   
        tag?: string;

        /**
         * The description of the items
         */ 
        description: {
            label?: string;
            value: string;
        }[]
    })[];
}
