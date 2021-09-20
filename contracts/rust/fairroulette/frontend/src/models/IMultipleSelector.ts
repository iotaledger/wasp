import type {ISelector} from './ISelector';

export interface IMultipleSelector extends ISelector {
    /**
     * The selectable values of the selector.
     */
     values: string[];

     /**
      * What hapen when clicking in an option
      */
     onClick: (indexSelected: number)=> void;
}
