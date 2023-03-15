import { writable } from 'svelte/store'
import type { IPopup } from './interfaces'

export const popupStore = writable<IPopup[]>([])
