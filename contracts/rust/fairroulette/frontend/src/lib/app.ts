import { get } from 'svelte/store';
import config from '../../config.dev';
import type { Bet } from './fairroulette_client';
import { FairRouletteService } from './fairroulette_client';
import { address, addressIndex, balance, fundsRequested, isWorking, keyPair, newAddressNeeded, placingBet, requestingFunds, round, seed, timestamp } from './store';
import { log } from './utils';
import {
    BasicClient, Colors, PoWWorkerManager,
    WalletService
} from './wasp_client';
import { Base58 } from './wasp_client/crypto/base58';
import { Seed } from './wasp_client/crypto/seed';

let client: BasicClient;
let walletService: WalletService;
let fairRouletteService: FairRouletteService;

let fundsUpdaterHandle;

const powManager: PoWWorkerManager = new PoWWorkerManager();

export async function initialize() {
    log('Page', 'Loading');

    if (config.seed) {
        seed.set(Base58.decode(config.seed));
    } else {
        seed.set(Seed.generate());
    }

    client = new BasicClient({
        GoShimmerAPIUrl: config.goshimmerApiUrl,
        WaspAPIUrl: config.waspApiUrl,
        SeedUnsafe: get(seed),
    });

    fairRouletteService = new FairRouletteService(client, config.chainId);
    walletService = new WalletService(client);

    powManager.load('/build/pow.worker.js');

    subscribeToRouletteEvents();
    setAddress(get(addressIndex));
    updateFunds();

    startFundsUpdater();

    // The best solution would be to call all of them in parallel. This is currently not possible.
    // As those requests can fail in certain cases, we need to wrap them in exception handlers,
    // to make sure that the other requests are being sent and that the page properly loads.
    const requests = [
        () =>
            fairRouletteService
                .getRoundStatus()
                .then((x) => round.update((_round) => {
                    _round.active = (x == 1);
                    return _round;
                })),
        () =>
            fairRouletteService
                .getRoundNumber()
                .then((x) => round.update((_round) => {
                    _round.number = x;
                    return _round;
                })),
        () =>
            fairRouletteService
                .getLastWinningNumber()
                .then((x) => round.update((_round) => {
                    _round.winningNumber = x;
                    return _round;
                })),
        () =>
            fairRouletteService.getRoundStartedAt().then((x) => round.update((_round) => {
                _round.startedAt = x;
                return _round;
            })),
    ];

    for (let request of requests) {
        await request().catch((e) => log('Error', e.message));
    }

    log('Page', 'Loaded');
}

export function setAddress(index: number) {
    addressIndex.set(index);

    address.set(Seed.generateAddress(get(seed), get(addressIndex)));
    keyPair.set(Seed.generateKeyPair(get(seed), get(addressIndex)));
}

export function createNewAddress() {
    addressIndex.set(get(addressIndex) + 1);
    setAddress(get(addressIndex));
}

export async function updateFunds() {
    let _balance = 0n;
    try {
        timestamp.set(Date.now() / 1000);
        _balance = await walletService.getFunds(
            get(address),
            Colors.IOTA_COLOR_STRING
        );
    } catch (ex) { }
    balance.set(_balance);

    log('Develop', `balance: ${get(balance)}, fundsRequested: ${get(fundsRequested)}`)
    if (get(balance) === 0n && get(fundsRequested) && get(newAddressNeeded)) {
        log('Develop', 'Create ner address')
        createNewAddress();
        newAddressNeeded.set(false)
    }
}

export function startFundsUpdater() {
    if (fundsUpdaterHandle) {
        fundsUpdaterHandle = clearInterval(fundsUpdaterHandle);
    }

    fundsUpdaterHandle = setInterval(updateFunds, 1000);
}

export async function placeBet() {
    newAddressNeeded.set(true);
    placingBet.set(true)
    isWorking.set(true);
    try {
        await fairRouletteService.placeBetOnLedger(
            get(keyPair),
            get(address),
            get(round).betSelection,
            get(round).betAmount,
        );
    } catch (ex) {
        log("Unknown", ex.message);

        throw ex;
    }
    isWorking.set(false);
}

export async function sendFaucetRequest() {
    requestingFunds.set(true);

    const faucetRequestResult = await walletService.getFaucetRequest(get(address));

    // In this example a difficulty of 20 is enough, might need a retune for prod to 21 or 22
    faucetRequestResult.faucetRequest.nonce =
        await powManager.requestProofOfWork(20, faucetRequestResult.poWBuffer);

    try {
        await client.sendFaucetRequest(faucetRequestResult.faucetRequest);
    } catch (ex) {
        log('Round', ex.message);
    }
    requestingFunds.set(false);
}

// To make sure the function gets called every second, we require that date.Now() is put in as a parameter to rely on sveltes change listener.
export function calculateRoundLengthLeft(timestamp: number) {
    const roundStarted = get(round).startedAt;

    if (roundStarted == 0) {
        return 0;
    }

    const diff = timestamp - roundStarted;

    // TODO: Explain.
    const executionCompensation = 5;
    const roundTimeLeft = Math.round(
        fairRouletteService.roundLength + executionCompensation - diff
    );

    if (roundTimeLeft <= 0) {
        return 0;
    }
    return roundTimeLeft;
}

export function subscribeToRouletteEvents() {
    fairRouletteService.on('roundStarted', (timestamp) => {
        round.update((_round) => {
            _round.active = true;
            _round.startedAt = timestamp;
            return _round;
        })
        log('Round', 'Started');
    });

    fairRouletteService.on('roundStopped', () => {
        round.update((_round) => {
            _round.active = false;
            return _round;
        })
        log('Round', 'Ended');
    });

    fairRouletteService.on('roundNumber', (roundNumber: bigint) => {
        round.update((_round) => {
            _round.number = roundNumber;
            return _round;
        })
        log('Round', `Current round number: ${roundNumber}`);
    });

    fairRouletteService.on('winningNumber', (winningNumber: bigint) => {
        round.update((_round) => {
            _round.winningNumber = winningNumber;
            return _round;
        })
        log('Round', `The winning number was: ${winningNumber}`);
    });

    fairRouletteService.on('betPlaced', (bet: Bet) => {
        placingBet.set(false);
        round.update((_round) => {
            _round.players.push(
                {
                    address: bet.better,
                    bet: bet.amount,
                },
            );
            return _round;
        })
        log(
            'Bet',
            `Bet placed from ${bet.better} on ${bet.betNumber} with ${bet.amount}`
        );
    });

    fairRouletteService.on('payout', (bet: Bet) => {
        log('Win', `Payout for ${bet.better} with ${bet.amount}`);
    });
}

export function isBroke(balance: bigint): boolean {
    return balance < 200;
}

export function isWealthy(balance: bigint): boolean {
    return balance > 200;
}