import { Arguments, ClientFunc, RequestID } from "..";
import { EventHandlers, Service } from "../service";
import { ServiceClient } from "../serviceclient";

export class AccountsService extends Service {
    public constructor(cl: ServiceClient) {
        const emptyEventHandlers: EventHandlers = new Map();
        super(cl, 0x3c4b5e02, emptyEventHandlers);
    }

    public deposit(): DepositFunc {
        return new DepositFunc(this);
    }
}

export class DepositFunc extends ClientFunc {
    private ArgAddress: string = "address";

    private args: Arguments = new Arguments();

    public address(v: string): void {
        this.args.setString(this.ArgAddress, v);
    }

    public async post(): Promise<RequestID> {
        return await super.post(0xbdc9102d, this.args, false);
    }
}
