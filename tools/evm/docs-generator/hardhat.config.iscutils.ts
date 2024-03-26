import { HardhatUserConfig } from "hardhat/config";
import 'solidity-docgen';
import path from 'path';

const rootPath = '../';
const cachePath = path.relative(path.resolve(rootPath), path.resolve('./cache'));
const artifactsPath = path.relative(path.resolve(rootPath), path.resolve('./artifacts'));
const outputDirPath = path.relative(path.resolve(rootPath), path.resolve('./docs/iscutils'));
const templatesPath = path.relative(path.resolve(rootPath), path.resolve('./templates'));

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  paths: {
    sources: 'iscutils',
    root: rootPath,
    cache: cachePath,
    artifacts: artifactsPath,
  },
  docgen: {
    outputDir: outputDirPath,
    templates: templatesPath,
    pages: 'files',
    exclude: [
      // Ignore test
      'prng_test.sol'
    ],
  },
};

export default config;

