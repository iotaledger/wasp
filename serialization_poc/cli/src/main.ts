'use strict';

import { SuiClient, } from '@mysten/sui.js/client';
import { Ed25519Keypair } from '@mysten/sui.js/keypairs/ed25519';
import { dirname, join, } from 'path';
import { ISCAnchorContract, publishISCContract, } from './contracts/isc';
import { prettyPrint } from './utils';
import { fileURLToPath } from 'url';
import { KEYPAIR_SECRET, SUI_HOST } from './consts';
import { doFaucetRequest } from './faucet';

async function main() {
  const keyPair = Ed25519Keypair.fromSecretKey(KEYPAIR_SECRET);
  const client = new SuiClient({ url: SUI_HOST });

  await doFaucetRequest(client, keyPair);

  const basePathISC = join(fileURLToPath(dirname(import.meta.url)), "../../../../kinesis/dapps/isc/");
  const iscContractDeploymentResult = await publishISCContract(client, keyPair, basePathISC);

  const iscContract = new ISCAnchorContract(client, keyPair, iscContractDeploymentResult);
  const chain = await iscContract.start_new_chain();

  await iscContract.send_request(chain.anchorId, {
    targetContract: "bank",
    entryPoint: "close_all_accounts_send_money_to_me",
    params: [[1, 0, 7, 4], [1, 3, 3, 7]],
  }, 1000n, null)

  prettyPrint(chain);
}


main();