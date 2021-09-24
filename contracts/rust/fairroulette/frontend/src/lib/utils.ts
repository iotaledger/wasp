export const generateRandomInt = (min: number = 0, max: number = 7, excluded: number = undefined): number => {
  let randomInt = Math.floor(Math.random() * (max - min + 1)) + min;
  return randomInt === excluded ? generateRandomInt(min, max, excluded) : randomInt;
}
