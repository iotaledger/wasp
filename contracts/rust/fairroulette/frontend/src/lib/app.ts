import { get } from 'svelte/store';
import config from '../../config.dev';
import type { Bet } from './fairroulette_client';
import { FairRouletteService } from './fairroulette_client';
import { Notification, NOTIFICATION_TIMEOUT_NEVER, showNotification } from './notifications';
import { address, addressIndex, balance, keyPair, placingBet, requestBet, requestingFunds, resetRound, round, seed, showWinnerAnimation, showWinningNumber, timestamp } from './store';
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
export const BETTING_NUMBERS = 8
export const ROUND_LENGTH = 60 //in seconds

enum LogTag {
    Page = 'Page',
    Round = 'Round',
    Win = 'Win',
    Error = 'Error',
    Unknown = 'Unknown'
}

export enum BettingStep {
    NumberChoice = 1,
    AmountChoice = 2,
}

export function log(tag: string, description: string) {
    round.update((_round) => {
        _round.logs.push({
            tag,
            description,
            timestamp: new Date().toLocaleTimeString(),
        },
        );
        return _round;
    })
}

export async function initialize() {
    log(LogTag.Page, 'Loading');

    if (config.seed) {
        seed.set(Base58.decode(config.seed));
    } else {
        seed.set(Seed.generate());
    }

    //TODO: Remove this at some point.
    if (!config.chainId && config.chainResolverUrl) {
        try {
            const response = await fetch(config.chainResolverUrl);
            const content = await response.json();
            config.chainId = content.chainId;
        } catch (ex) {
            showNotification({
                type: Notification.Error,
                message: ex.message,
                timeout: NOTIFICATION_TIMEOUT_NEVER
            })
            log(LogTag.Error, ex.message);
        }
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
                .then((x) => round.update($round => ({ ...$round, active: x == 1 }))),
        () =>
            fairRouletteService
                .getRoundNumber()
                .then((x) => round.update($round => ({ ...$round, number: x }))),
        () =>
            fairRouletteService
                .getLastWinningNumber()
                .then((x) => round.update($round => ({ ...$round, winningNumber: x }))),
        () =>
            fairRouletteService
                .getRoundStartedAt()
                .then((x) => round.update($round => ({ ...$round, startedAt: x }))),
    ];

    for (let request of requests) {
        await request().catch((e) => log(LogTag.Error, e.message));
    }

    log(LogTag.Page, 'Loaded');
}

export function setAddress(index: number) {
    addressIndex.set(index);

    address.set(Seed.generateAddress(get(seed), get(addressIndex)));
    keyPair.set(Seed.generateKeyPair(get(seed), get(addressIndex)));
}

export function createNewAddress() {
    addressIndex.update(($addressIndex) => $addressIndex + 1);
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
}

export function startFundsUpdater() {
    if (fundsUpdaterHandle) {
        fundsUpdaterHandle = clearInterval(fundsUpdaterHandle);
    }

    fundsUpdaterHandle = setInterval(updateFunds, 1000);
}

export async function placeBet() {
    placingBet.set(true)
    showWinningNumber.set(false)
    try {
        await fairRouletteService.placeBetOnLedger(
            get(keyPair),
            get(address),
            get(round).betSelection,
            get(round).betAmount,
        );
    } catch (ex) {
        showNotification({
            type: Notification.Error,
            title: 'Error placing bet',
            message: ex.message,
            timeout: NOTIFICATION_TIMEOUT_NEVER
        })

        log(LogTag.Unknown, ex.message);

        throw ex;
    }
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
        showNotification({
            type: Notification.Error,
            message: ex.message,
            timeout: NOTIFICATION_TIMEOUT_NEVER
        })

        log(LogTag.Round, ex.message);
    }
    requestingFunds.set(false);
}

export function calculateRoundLengthLeft(timestamp: number) {
    const roundStarted = get(round).startedAt;

    if (!timestamp || !roundStarted) return undefined

    if (roundStarted == 0) {
        return 0;
    }

    const diff = timestamp - roundStarted;

    // TODO: Explain.
    const executionCompensation = 5;
    const roundTimeLeft = Math.round(
        fairRouletteService?.roundLength + executionCompensation - diff
    );

    if (roundTimeLeft <= 0) {
        return 0;
    }
    return roundTimeLeft;
}

export function subscribeToRouletteEvents() {
    fairRouletteService.on('roundStarted', (timestamp) => {
        showWinningNumber.set(false);
        round.update($round => ({ ...$round, active: true, startedAt: timestamp, logs: [] }))
        log(LogTag.Round, 'Started');
    });

    fairRouletteService.on('roundStopped', () => {
        if (get(placingBet) || get(requestBet)) {
            showNotification({
                type: Notification.Info,
                message: 'The current round just ended. Your bet will be placed in the next round. ',
                timeout: NOTIFICATION_TIMEOUT_NEVER
            })
        }
        resetRound();
        log(LogTag.Round, 'Ended');
    });

    fairRouletteService.on('roundNumber', (roundNumber: bigint) => {
        round.update($round => ({ ...$round, number: roundNumber }))
        log(LogTag.Round, `Current round number: ${roundNumber}`);
    });

    fairRouletteService.on('winningNumber', (winningNumber: bigint) => {
        round.update($round => ({ ...$round, winningNumber }))
        showWinningNumber.set(true);
        log(LogTag.Round, `The winning number was: ${winningNumber}`);
    });

    fairRouletteService.on('betPlaced', (bet: Bet) => {
        placingBet.set(false);
        round.update(($round) => {
            if (bet.better === get(address)) {
                $round.betPlaced = true;
                $round.betAmount = 0n;
            }
            $round.players.push(
                {
                    address: bet.better,
                    bet: bet.amount,
                },
            );
            return $round;
        })
        log(
            'Bet',
            `Bet placed from ${bet.better} on ${bet.betNumber} with ${bet.amount}`
        );
        requestBet.set(false);
    });

    fairRouletteService.on('payout', (bet: Bet) => {
        if (bet.better === get(address)) {
            showNotification({
                type: Notification.Win,
                message: `Congratulations! You just won the round. You received ${bet.amount} iotas.`,
                timeout: NOTIFICATION_TIMEOUT_NEVER
            })
            showWinnerAnimation();
        }
        else if (get(round).betPlaced) {
            showNotification({
                type: Notification.Info,
                message: 'Sorry, you have lost. Try your luck again.',
                timeout: NOTIFICATION_TIMEOUT_NEVER
            })
        }
        log(LogTag.Win, `Payout for ${bet.better} with ${bet.amount}`);
    });
}

export function isWealthy(balance: bigint): boolean {
    return balance >= 200;
}