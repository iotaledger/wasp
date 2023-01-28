import {SingleNodeClient} from '@iota/iota.js';

export class ClientLib {
    private nodeUrl: string;
    private nodeClient: SingleNodeClient;

    constructor(nodeUrl: string) {
        this.nodeUrl = nodeUrl;
        this.nodeClient = new SingleNodeClient(this.nodeUrl);
    }

    public async isHealthy(): Promise<boolean> {
        return await this.nodeClient.health();
    }
}