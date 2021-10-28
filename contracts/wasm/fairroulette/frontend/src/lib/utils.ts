export const generateRandomInt = (
  min: number = 0,
  max: number = 7,
  excluded: number | undefined = undefined,
): number => {
  const randomInt = Math.floor(Math.random() * (max - min + 1)) + min;
  return randomInt === excluded ? generateRandomInt(min, max, excluded) : randomInt;
};

export const generateRandomId = (): string => {
  const randomNumber = crypto.getRandomValues(new Uint8Array(16));

  return Array.from(randomNumber, (byte) => {
    return ('0' + (byte & 0xff).toString(16)).slice(-2);
  }).join('');
};

export const delay = (ms: number): Promise<void> => {
  return new Promise((resolve) => {
    // eslint-disable-next-line @typescript-eslint/no-implied-eval
    setTimeout(resolve, ms);
  });
};

declare global {
  interface Window {
    dataLayer: unknown[];
  }
}

export const loadGoogleAnalytics = (gaID: string): void => {
  window.dataLayer = window.dataLayer || [];
  function gtag(key: string, value: unknown) {
    // eslint-disable-next-line prefer-rest-params
    window.dataLayer.push(arguments);
  }

  gtag('js', new Date());
  gtag('config', gaID);

  const script = document.createElement('script');
  script.src = `https://www.googletagmanager.com/gtag/js?id=${gaID}`;
  document.body.appendChild(script);
};

/*
 * General utils for managing cookies in Typescript.
 * Source: https://gist.github.com/joduplessis/7b3b4340353760e945f972a69e855d11
 */

export const setCookie = (name: string, val: string, expDays: number): void => {
  const date = new Date();
  const value = val;
  date.setTime(date.getTime() + expDays * 24 * 60 * 60 * 1000);
  document.cookie = name + '=' + value + '; expires=' + date.toUTCString() + '; path=/';
};
export function getCookie(name: string): string | undefined {
  const value = '; ' + document.cookie;
  const parts = value.split('; ' + name + '=');

  if (parts?.length == 2) {
    return parts?.pop()?.split(';')?.shift() ?? undefined;
  }
}
