
export interface IButton {
    /**
     * The label.
     */
    label: string;

    /**
     * Is the button disabled.
     */
    disabled?: boolean;

    /**
     * What to do when button is clicked.
     */
     onClick(): void;
}
