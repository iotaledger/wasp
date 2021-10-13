import { get } from 'svelte/store';
import config from '../../config.dev';
import type { Bet } from './fairroulette_client';
import { FairRouletteService } from './fairroulette_client';
import { Notification, showNotification } from './notifications';
import { address, addressesHistory, addressIndex, balance, firstTimeRequestingFunds, isAWinnerPlayer, keyPair, placingBet, receivedRoundStarted, requestingFunds, resetRound, round, seed, showBettingSystem, showWinnerAnimation, showWinningNumber, timestamp } from './store';
import {
    BasicClient, Colors, PoWWorkerManager,
    WalletService
} from './wasp_client';
import { Base58 } from './wasp_client/crypto/base58';
import { Seed } from './wasp_client/crypto/seed';

let client: BasicClient;
let walletService: WalletService;
let fairRouletteService: FairRouletteService;

let fundsUpdaterHandle: NodeJS.Timer | undefined;

const powManager: PoWWorkerManager = new PoWWorkerManager();
export const BETTING_NUMBERS = 8
export const ROUND_LENGTH = 60 //in seconds

const DEFAULT_AUTODISMISS_TOAST_TIME = 5000 //in milliseconds

export enum LogTag {
    Site = 'Site',
    Round = 'Round',
    Funds = 'Funds',
    SmartContract = 'Smart Contract',
    Error = 'Error'
}

export enum BettingStep {
    NumberChoice = 1,
    AmountChoice = 2,
}

export enum StateMessage {
    Running = 'Running',
    Start = 'Start',
    AddFunds = 'AddFunds',
    ChoosingNumber = 'ChoosingNumber',
    ChoosingAmount = 'ChoosingAmount',
    PlacingBet = 'PlacingBet'
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
    log(LogTag.Site, 'Initializing wallet.');

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
        } catch (ex: any) {
            showNotification({
                type: Notification.Error,
                message: ex.message,
                timeout: DEFAULT_AUTODISMISS_TOAST_TIME
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

    log(LogTag.Site, 'Demo loaded');
}

export function setAddress(index: number) {
    addressIndex.set(index);

    address.set(Seed.generateAddress(get(seed), get(addressIndex)));
    keyPair.set(Seed.generateKeyPair(get(seed), get(addressIndex)));
}

export function createNewAddress() {
    addressesHistory.update(_history => [..._history, get(address)])
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
    } catch (ex: any) { }
    balance.set(_balance);
}

export function startFundsUpdater() {
    if (fundsUpdaterHandle) {
        clearInterval(fundsUpdaterHandle);
        fundsUpdaterHandle = undefined;
    }
    fundsUpdaterHandle = setInterval(updateFunds, 1000);
}

export async function placeBet() {
    placingBet.set(true)
    showBettingSystem.set(false)
    showWinningNumber.set(false)
    try {
        await fairRouletteService.placeBetOnLedger(
            get(keyPair),
            get(address),
            get(round).betSelection,
            get(round).betAmount,
        );
    } catch (ex: any) {
        showNotification({
            type: Notification.Error,
            title: 'Error placing bet',
            message: ex.message,
            timeout: DEFAULT_AUTODISMISS_TOAST_TIME
        })

        log(LogTag.Error, ex.message);

        throw ex;
    }
}

export async function sendFaucetRequest() {
    log(LogTag.Funds, "Funds requested from devnet. The GoShimmer nodes have received them.  Sending funds to the requested wallet.");

    if (!get(firstTimeRequestingFunds)) {
        createNewAddress();
    }
    firstTimeRequestingFunds.set(false);

    requestingFunds.set(true);

    const faucetRequestResult = await walletService.getFaucetRequest(get(address));

    // In this example a difficulty of 20 is enough, might need a retune for prod to 21 or 22
    faucetRequestResult.faucetRequest.nonce =
        await powManager.requestProofOfWork(20, faucetRequestResult.poWBuffer);

    try {
        await client.sendFaucetRequest(faucetRequestResult.faucetRequest);
        log(LogTag.SmartContract, "Funds sent to Wasp chain address using GoShimmer nodes");
    } catch (ex) {
        showNotification({
            type: Notification.Error,
            message: ex.message,
            timeout: DEFAULT_AUTODISMISS_TOAST_TIME
        })

        log(LogTag.Error, ex.message);
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
        receivedRoundStarted.set(true);
        showWinningNumber.set(false);
        round.update($round => ({ ...$round, active: true, startedAt: timestamp, logs: [] }))
    });

    fairRouletteService.on('roundStopped', () => {
        if (get(placingBet) || get(showBettingSystem)) {
            showNotification({
                type: Notification.Info,
                message: 'The current round just ended. Your bet will be placed in the next round. ',
                timeout: DEFAULT_AUTODISMISS_TOAST_TIME
            })
        }
        else if (get(round).betPlaced && !get(isAWinnerPlayer)) {
            showNotification({
                type: Notification.Info,
                message: 'Sorry, you lost this round. Try again!',
                timeout: DEFAULT_AUTODISMISS_TOAST_TIME
            })
        }

        resetRound();
        log(LogTag.Round, 'Ended');
        log(LogTag.SmartContract, "Round Ended. Current bets cleared.");
    });

    fairRouletteService.on('roundNumber', (roundNumber: bigint) => {
        round.update($round => ({ ...$round, number: roundNumber }))
        log(LogTag.Round, `Started. Round number: ${roundNumber}`);
    });

    fairRouletteService.on('winningNumber', (winningNumber: bigint) => {
        round.update($round => ({ ...$round, winningNumber }))
        showWinningNumber.set(true);

        log(LogTag.SmartContract, "The winning number was decided");
        log(LogTag.SmartContract, `${winningNumber} is the winning number!`);
    });

    fairRouletteService.on('betPlaced', (bet: Bet) => {
        placingBet.set(false);
        round.update(($round) => {
            if (bet.better === get(address)) {
                $round.betPlaced = true;
                $round.betAmount = 0n;
                log(LogTag.SmartContract, "Your number and betting amounts are saved");
            }
            $round.players.push(
                {
                    address: bet.better,
                    bet: bet.amount,
                    number: bet.betNumber
                },
            );
            return $round;
        })

    });

    fairRouletteService.on('payout', (bet: Bet) => {

        if (bet.better === get(address) || get(addressesHistory).includes(bet.better)) {
            showNotification({
                type: Notification.Win,
                message: `Congratulations! You just won the round. You received ${bet.amount} iotas.`,
                timeout: DEFAULT_AUTODISMISS_TOAST_TIME
            })
            showWinnerAnimation();
        }
        log(LogTag.SmartContract, `Payout for ${bet.better} with ${bet.amount}`);
    });
}

export function isWealthy(balance: bigint): boolean {
    return balance >= 200;
}