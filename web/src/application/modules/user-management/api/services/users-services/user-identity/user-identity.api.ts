import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type { UserIdentityApiValidator, UserIdentity_FindMine_Res } from "~/user-management/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class UserIdentityApi extends BaseApi {
    public constructor(private readonly validator: UserIdentityApiValidator) {
        super();
    }

    /**
     * Find the signed-in person's identity
     */
    async findMine(signal?: AbortSignal): Promise<Result<UserIdentity_FindMine_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/users/current/identity", {
                    signal,
                }),
            ).pipe(
                map(this.validator.findMine),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
