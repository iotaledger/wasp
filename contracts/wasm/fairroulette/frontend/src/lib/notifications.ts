import { writable } from 'svelte/store';
import { generateRandomId } from './utils';

const NOTIFICATION_TIMEOUT_DEFAULT = 5000;
export const NOTIFICATION_TIMEOUT_NEVER = -1;

export enum Notification {
  Win = 'win',
  Error = 'error',
  Info = 'info',
  Lose = 'lose',
}
export type NotificationData = {
  type: Notification;
  title?: string;
  message: string;
  id?: string;
  timeout?: number;
};

export const displayNotifications = writable<Array<NotificationData>>([]);

export function showNotification(notificationData: NotificationData): string {
  notificationData.id = generateRandomId();
  notificationData.timeout = notificationData.timeout ?? NOTIFICATION_TIMEOUT_DEFAULT;

  displayNotifications.update((_currentNotifications) => {
    _currentNotifications.push(notificationData);
    return _currentNotifications;
  });

  if (notificationData.timeout !== NOTIFICATION_TIMEOUT_NEVER) {
    setTimeout(() => removeDisplayNotification(notificationData.id), notificationData.timeout);
  }

  return notificationData.id;
}

export function removeDisplayNotification(id: string | undefined): void {
  displayNotifications.update((_currentNotifications) => {
    const idx = _currentNotifications.findIndex((n) => n.id === id);
    if (idx >= 0) {
      _currentNotifications.splice(idx, 1);
    }
    return _currentNotifications;
  });
}

export function updateDisplayNotification(id: string, updateData: NotificationData): void {
  displayNotifications.update((_currentNotifications) => {
    const notification = _currentNotifications.find((n) => n.id === id);
    if (notification) {
      notification.message = updateData.message;
      notification.timeout = updateData.timeout ?? NOTIFICATION_TIMEOUT_DEFAULT;

      if (notification.timeout !== NOTIFICATION_TIMEOUT_NEVER) {
        setTimeout(() => removeDisplayNotification(notification.id), notification.timeout);
      }
    }
    return _currentNotifications;
  });
}
