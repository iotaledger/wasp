
export const randomIntFromInterval = (min = 0, max = 5) => {
  return Math.floor(Math.random() * (max - min + 1) + min);
}