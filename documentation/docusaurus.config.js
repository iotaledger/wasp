const path = require('path');

module.exports = {
  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'wasp',
        path: path.resolve(__dirname, 'docs'),
        routeBasePath: 'smart-contracts',
        sidebarPath: path.resolve(__dirname, 'sidebars.js'),
        editUrl: 'https://github.com/iotaledger/wasp/edit/develop/documentation',
        remarkPlugins: [require('remark-code-import'), require('remark-import-partial'), require('remark-remove-comments')],
        versions: {
          current: {
            label: 'Develop',
            path: 'develop'
          },
        },
      }
    ],
  ],
  staticDirectories: [path.resolve(__dirname, 'static')],
};
