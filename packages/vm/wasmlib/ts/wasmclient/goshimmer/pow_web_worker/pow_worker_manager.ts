import { v4 as uuidv4 } from 'uuid';
import type { Buffer } from '../../buffer';
import Worker from 'worker-loader!./pow.worker.ts';

export interface PowWorkerRequest {
    type: string;
    data: Buffer;
    uuid: string;
    difficulty: number;
}

export interface PowWorkerResponse {
    type: string;
    data: number;
    uuid: string;
    error?: Error;
}

export class PoWWorkerManager {
    private powWorker: Worker = new Worker();

    public requestProofOfWork(difficulty: number, data: Buffer): Promise<number> {
        return new Promise((resolve, reject) => {
            if (this.powWorker == null) throw new Error('powWorker not defined');

            const requestId = uuidv4();

            const responseHandler = (e: MessageEvent) => {
                const message: PowWorkerResponse = e.data;

                if (message.type == 'pow_response' && message.uuid == requestId) {
                    this.powWorker!.removeEventListener('message', responseHandler);

                    if (!message.error) {
                        resolve(message.data);
                    } else {
                        reject(message.error);
                    }
                }
            };

            this.powWorker.addEventListener('message', responseHandler);

            const request: PowWorkerRequest = {
                type: 'pow_request',
                data: data,
                difficulty: difficulty,
                uuid: requestId,
            };

            this.powWorker.postMessage(request);
        });
    }
}
