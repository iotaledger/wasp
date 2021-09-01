# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern 
static website generator.

## Installation

```console
yarn install
```

## Local Development

```console
yarn start
```

This command starts a local development server and opens up a browser window. 
Most changes are reflected live without having to restart the server.

## Build

```console
yarn build
```

This command generates static content into the `build` directory and can 
be served using any static contents hosting service.

## Deployment

There is an automatic deployment flow in place which will automatically trigger 
on a merged PR to the `master` branch
To deploy in a local fork make sure you update `docusaurus.config.js` `baseUrl` 
setting to `/wasp/` temporarily and merge this to your master branch. 
This will trigger the workflow (make sure you have github pages enabled on the 
`gh-pages` branch in your settings.
Don't commit this `baseUrl` setting to production, it needs to be `/` in the 
main repo to use the CNAME provided.
