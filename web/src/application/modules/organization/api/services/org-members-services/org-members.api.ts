import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    OrgMembers_FindMany_Req,
    OrgMembers_FindMany_Res,
    OrgMembers_UpdateRole_Req,
    OrgMembers_UpdateRole_Res,
    OrgMembers_UpdateStatus_Req,
    OrgMembers_UpdateStatus_Res,
} from "./org-members.api.contracts";
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

    /**
     * Make a member an owner, an admin or a member. Only an owner makes an owner or changes one; the last active owner
     * stays one (ERR_ORG_LAST_OWNER). A removed member is not found.
     */
    async updateRole(request: OrgMembers_UpdateRole_Req): Promise<Result<OrgMembers_UpdateRole_Res, Error>> {
        const { orgId, userId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/members/${userId}/role`, payload)).pipe(
                map(this.validator.updateRole),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Suspend a member, remove them, or make them active again. Anybody may remove themselves: that is leaving.
     * Removing takes their role and SSH accounts away; the last active owner stays one (ERR_ORG_LAST_OWNER).
     */
    async updateStatus(request: OrgMembers_UpdateStatus_Req): Promise<Result<OrgMembers_UpdateStatus_Res, Error>> {
        const { orgId, userId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/members/${userId}/status`, payload)).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
