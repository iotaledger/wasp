export function handleEnterKeyDown(
  event: KeyboardEvent,
  callback: () => void,
): void {
  if (event?.key === 'Enter' && callback && typeof callback === 'function') {
    callback();
  }
}

export function handleEscapeKeyDown(
  event: KeyboardEvent,
  callback: () => void,
): void {
  if (event?.key === 'Escape' && callback && typeof callback === 'function') {
    callback();
  }
}

export const truncateText = (
  text: string,
  charsStart: number = 6,
  charsEnd: number = 4,
) =>
  text?.length > charsStart + charsEnd
    ? `${text?.slice(0, charsStart)}...${text?.slice(-charsEnd)}`
    : text;

export const copyToClipboard = (content: string): Promise<void> =>
  navigator.clipboard.writeText(content);
