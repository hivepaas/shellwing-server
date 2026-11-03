import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SshPrincipals_FindOne_Req,
    SshPrincipals_FindOne_Res,
    SshPrincipals_UpdateOne_Req,
    SshPrincipals_UpdateOne_Res,
} from "./ssh-principals.api.contracts";
import type { SshPrincipalsApiValidator } from "./ssh-principals.api.validator";

export class SshPrincipalsApi extends BaseApi {
    public constructor(private readonly validator: SshPrincipalsApiValidator) {
        super();
    }

    /**
     * The accounts a member's certificates may log in as
     */
    async findOne(
        request: SshPrincipals_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SshPrincipals_FindOne_Res, Error>> {
        const { orgId, userId } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/orgs/${orgId}/ssh/members/${userId}/principals`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Replace the accounts a member's certificates may log in as, from their next certificate on
     */
    async updateOne(request: SshPrincipals_UpdateOne_Req): Promise<Result<SshPrincipals_UpdateOne_Res, Error>> {
        const { orgId, userId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/ssh/members/${userId}/principals`, payload)).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
