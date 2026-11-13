import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SshKrlToken_CreateOne_Req,
    SshKrlToken_CreateOne_Res,
    SshKrlToken_DeleteOne_Req,
    SshKrlToken_DeleteOne_Res,
} from "./ssh-krl.api.contracts";
import type { SshKrlApiValidator } from "./ssh-krl.api.validator";

export class SshKrlApi extends BaseApi {
    public constructor(private readonly validator: SshKrlApiValidator) {
        super();
    }

    /**
     * Make a token the organization's servers fetch the revocation list with. One made before stops working, so this
     * is also how a token that may have leaked is replaced.
     */
    async createOne(request: SshKrlToken_CreateOne_Req): Promise<Result<SshKrlToken_CreateOne_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/orgs/${orgId}/ssh/krl/token`)).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Take the token away. The servers holding it stop being able to fetch; the list each has in place keeps working.
     */
    async deleteOne(request: SshKrlToken_DeleteOne_Req): Promise<Result<SshKrlToken_DeleteOne_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/orgs/${orgId}/ssh/krl/token`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
