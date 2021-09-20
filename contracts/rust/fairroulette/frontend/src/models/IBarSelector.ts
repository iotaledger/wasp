import type {ISelector} from './ISelector';

export interface IBarSelector extends ISelector {
    /**
     * The minimum selectable option in the range.
     */
     minimum: number;

    /**
     * The maximum selectable option in the range.
     */
     maximum: number;

    /**
     * The step between two adjacent values in the range.
     */
     step?: number;

    /**
     * The unit symbol of the selectable values.
     */
     unit?: string;

     /**
      * What hapen when clicking in an option
      */
     onChange: (value: number)=> void;
}
