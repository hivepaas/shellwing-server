import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { Orgs_FindMany_Res } from "./orgs.api.contracts";
import type { OrgsApiValidator } from "./orgs.api.validator";

export class OrgsApi extends BaseApi {
    public constructor(private readonly validator: OrgsApiValidator) {
        super();
    }

    /**
     * The organizations the signed-in user is a member of, with their role in each
     */
    async findMany(signal?: AbortSignal): Promise<Result<Orgs_FindMany_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/orgs", {
                    signal,
                }),
            ).pipe(
                map(this.validator.findMany),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
