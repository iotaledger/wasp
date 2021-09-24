import { round } from './store';

export function log(tag: string, description: string) {
  round.update((_round) => {
    _round.logs.push({
      tag: tag,
      description: description,
      timestamp: new Date().toLocaleTimeString(),
    },
    );
    return _round;
  })
}

export const generateRandomInt = (min: number = 0, max: number = 7, excluded: number = undefined): number => {
  let randomInt = Math.floor(Math.random() * (max - min + 1)) + min;
  return randomInt === excluded ? generateRandomInt(min, max, excluded) : randomInt;
}
