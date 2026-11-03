import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { OrgMembers_FindMany_Req, OrgMembers_FindMany_Res } from "./org-members.api.contracts";
import type { OrgMembersApiValidator } from "./org-members.api.validator";

export class OrgMembersApi extends BaseApi {
    public constructor(private readonly validator: OrgMembersApiValidator) {
        super();
    }

    /**
     * The organization's members, with the accounts their certificates may log in as
     */
    async findMany(
        request: OrgMembers_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<OrgMembers_FindMany_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/orgs/${orgId}/members`, {
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
