import { v4 as uuidv4 } from 'uuid';

export interface PowWorkerRequest {
  type: string
  data: any
  uuid: string
  difficulty: number
}

export interface PowWorkerResponse {
  type: string
  data: any
  uuid: string
  error?: Error
}


export class PoWWorkerManager {

  private powWorker: Worker;

  public Load(url: string) {
    this.powWorker = new Worker(url);
  }

  public RequestProofOfWork(difficulty: number, data: any): Promise<number> {
    return new Promise((resolve, reject) => {
      console.log("Requesting")
      const requestId = uuidv4();

      const responseHandler = (e) => {
        const message: PowWorkerResponse = e.data;

        console.log("Got response");
        console.log(message);

        if (message.type === 'pow_response' && message.uuid === requestId) {
          this.powWorker.removeEventListener('message', responseHandler);

          if (!message.error) {
            resolve(message.data);
          }
          else {
            reject(message.error);
          }
        }
      };

      this.powWorker.addEventListener('message', responseHandler);

      const request: PowWorkerRequest = { type: 'pow_request', data: data, difficulty: difficulty, uuid: requestId };

      this.powWorker.postMessage(request);

      console.log("Pow request sent");
      console.log(request)
    });
  }
}
