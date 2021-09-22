import type { ITypeBase } from "./ITypeBase";

/**
 * The global type for the entries panel.
 */
export const PLAYER_ENTRIES_TYPE = 0;

export type IPlayerData = {
    address: string;
    fields: {
        label: string;
        value: string;
    }[];
}
export interface IPlayerEntries extends ITypeBase<0> {
    data: IPlayerData[];
}
