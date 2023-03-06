import { writable } from 'svelte/store'
import type { NetworkOption } from './lib/network_option'

export const network = writable<NetworkOption>();