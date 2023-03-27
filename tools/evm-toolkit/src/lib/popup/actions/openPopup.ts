import type { PopupId } from '../enums'
import { popupStore } from '../store'
import type { PopupProps } from '../types'

export function openPopup(popupId: PopupId, props?: PopupProps): void {
    const newPopup = { id: popupId, props }
    popupStore.update(($popupStore) => [...$popupStore, newPopup])
}
