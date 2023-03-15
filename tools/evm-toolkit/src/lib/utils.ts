export function handleEnterKeyDown(event: KeyboardEvent, callback: () => void): void {
  if (event?.key === 'Enter' && callback && typeof callback === 'function') {
      callback()
  }
}

export function handleEscapeKeyDown(event: KeyboardEvent, callback: () => void): void {
  if (event?.key === 'Escape' && callback && typeof callback === 'function') {
      callback()
  }
}
