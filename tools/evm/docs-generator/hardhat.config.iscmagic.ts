import { HardhatUserConfig } from "hardhat/config";
import 'solidity-docgen';
import path from 'path';

const rootPath = '../../../packages/vm/core/evm/';
const cachePath = path.relative(path.resolve(rootPath), path.resolve('./cache'));
const artifactsPath = path.relative(path.resolve(rootPath), path.resolve('./artifacts'));
const outputDirPath = path.relative(path.resolve(rootPath), path.resolve('./docs/iscmagic'));
const templatesPath = path.relative(path.resolve(rootPath), path.resolve('./templates'));

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  paths: {
    sources: 'iscmagic',
    root: rootPath,
    cache: cachePath,
    artifacts: artifactsPath,
  },
  docgen: { 
    outputDir: outputDirPath,
    templates: templatesPath,
    pages: 'files',
  },
};

export default config;
