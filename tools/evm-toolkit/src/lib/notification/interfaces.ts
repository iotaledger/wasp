import type { NotificationType } from './enums'

export interface INotification {
    id: string
    message: string
    type: NotificationType
    duration?: number
}
