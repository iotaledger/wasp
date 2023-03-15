import type { PopupId } from './enums'
import type { PopupProps } from './types'

export interface IPopup {
    id: PopupId
    props?: PopupProps
}
