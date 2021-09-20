import type { ILogEntries } from "./ILogEntries";
import type { IPlayerEntries } from "./IPlayerEntries";
import type { ITypeBase } from "./ITypeBase";

/**
 * The global type for the entries panel.
 */
export const ENTRIES_PANEL_TYPE = 2;
export interface IEntriesPanel extends ITypeBase<2> {
    /**
     * The title of the panel.
     */
     title: string;

     /**
      * Is items ordered
      */
     ordered?: boolean;
    /**
     * The entries.
     */
    entries: (IPlayerEntries | ILogEntries);
}
