import { inspect } from 'util';

export const prettyPrint = (/** @type {Object} */ x, depth = 5) => console.log(inspect(x, { colors: true, depth: depth }));
export const delay = async (/** @type {number} */ ms) => new Promise((resolve) => setTimeout(resolve, ms));