import { SuiClient } from "@mysten/sui.js/client";
import { requestSuiFromFaucetV1 } from "@mysten/sui.js/faucet";
import { Ed25519Keypair } from "@mysten/sui.js/keypairs/ed25519";
import { MIST_PER_SUI } from "@mysten/sui.js/utils";
import { delay } from "./utils";
import { SUI_FAUCET_HOST } from "./consts";

async function parallelFaucetRequest(nTimes: number, address: string) {
  const requests = Array.from({ length: nTimes }, () =>
    requestSuiFromFaucetV1({
      host: SUI_FAUCET_HOST,
      recipient: address,
    })
  );

  await Promise.all(requests);
}

export async function doFaucetRequest(client: SuiClient, keyPair: Ed25519Keypair) {
  const balance = await client.getBalance({
    owner: keyPair.toSuiAddress(),
  });

  if (BigInt(balance.totalBalance) < MIST_PER_SUI * 10n) {
    const faucetIterations = 10;

    console.log(`Requesting funds ${faucetIterations} times.`);
    await parallelFaucetRequest(faucetIterations, keyPair.toSuiAddress());

    console.log('Waiting for funds to arrive');
    await delay(15000);
  }
}