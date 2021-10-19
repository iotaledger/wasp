export const generateRandomInt = (min: number = 0, max: number = 7, excluded: number | undefined = undefined): number => {
  let randomInt = Math.floor(Math.random() * (max - min + 1)) + min;
  return randomInt === excluded ? generateRandomInt(min, max, excluded) : randomInt;
}

export const generateRandomId = (): string => {
  return Array.from(crypto.getRandomValues(new Uint8Array(16)), (byte) => {
    return ('0' + (byte & 0xff).toString(16)).slice(-2)
  }).join('')
}

export const delay = (ms: number): Promise<void> => {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}