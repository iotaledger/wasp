export const generateRandomInt = (min: number = 0, max: number = 7, excluded: number | undefined = undefined): number => {
  let randomInt = Math.floor(Math.random() * (max - min + 1)) + min;
  return randomInt === excluded ? generateRandomInt(min, max, excluded) : randomInt;
}

export const generateRandomId = (): string => {
  return Array.from(crypto.getRandomValues(new Uint8Array(16)), (byte) => {
    return ('0' + (byte & 0xff).toString(16)).slice(-2)
  }).join('')
}

export const googleAnalytics = (gaID: string): void => {
  window.dataLayer = window.dataLayer || []
  function gtag() { dataLayer.push(arguments) }
  gtag('js', new Date())

  gtag('config', gaID)

  const script = document.createElement('script')
  script.src = `https://www.googletagmanager.com/gtag/js?id=${gaID}`
  document.body.appendChild(script)
}

export const setCookie = (cookieName: string, cookieValue: string | boolean, expirationDays: number): void => {
  var d = new Date()
  d.setTime(d.getTime() + expirationDays * 24 * 60 * 60 * 1000)
  var expires = 'expires=' + d.toUTCString()
  if (document) document.cookie = cookieName + '=' + cookieValue + ';' + expires + ';path=/'
}

export const getCookie = (cookieName: string): string => {
  if (document) {
    var name = cookieName + '='
    var ca = document.cookie.split(';')
    for (var i = 0; i < ca.length; i++) {
      var c = ca[i]
      while (c?.charAt(0) == ' ') {
        c = c?.substring(1)
      }
      if (c?.indexOf(name) == 0) {
        return c?.substring(name?.length, c?.length)
      }
    }
    return ''
  }
  return ''
}