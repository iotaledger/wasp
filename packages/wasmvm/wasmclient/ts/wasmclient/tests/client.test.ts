import { ClientLib } from '../lib/client';

const nodeUrl = 'https://chrysalis-nodes.iota.org';

describe('clientLib', function () {
    const clientLib = new ClientLib(nodeUrl);

    describe('get node health', function  () {
        it('should return node health', async () => {
            const isHealthy = await clientLib.isHealthy();
            expect(isHealthy).toBeTruthy();
        });
    });
});
