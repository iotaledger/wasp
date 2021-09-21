
import type { ITypeBase } from "./ITypeBase";

/**
 * The global type for the entries panel.
 */
export  const LOG_ENTRIES_TYPE = 1;

export type ILogData = {
    /**
     * The label above the title.
     */
    tag: string;

    /**
     * The timestamp.
     */
    timestamp: string;

    /**
     * The data fields of a player.
     */ 
    description: string;
}

export interface ILogEntries extends ITypeBase<1>{
    data: ILogData []
}
