import type { PopupId } from '../enums'
import { popupStore } from '../store'

export function closePopup(popupId: PopupId): void {
    popupStore.update(($popupStore) => $popupStore.filter((popup) => popup.id !== popupId))
}
