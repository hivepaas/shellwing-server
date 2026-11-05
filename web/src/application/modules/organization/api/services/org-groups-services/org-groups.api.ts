import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    OrgGroups_CreateOne_Req,
    OrgGroups_CreateOne_Res,
    OrgGroups_DeleteOne_Req,
    OrgGroups_DeleteOne_Res,
    OrgGroups_FindMany_Req,
    OrgGroups_FindMany_Res,
    OrgGroups_SetMembers_Req,
    OrgGroups_SetMembers_Res,
    OrgGroups_UpdateOne_Req,
    OrgGroups_UpdateOne_Res,
} from "./org-groups.api.contracts";
import type { OrgGroupsApiValidator } from "./org-groups.api.validator";

export class OrgGroupsApi extends BaseApi {
    public constructor(private readonly validator: OrgGroupsApiValidator) {
        super();
    }

    /**
     * The organization's groups, with their accounts and members
     */
    async findMany(
        request: OrgGroups_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<OrgGroups_FindMany_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/orgs/${orgId}/groups`, {
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
     * Make a group, with no members. A name taken already, whatever its case, is refused (ERR_ALREADY_EXIST).
     */
    async createOne(request: OrgGroups_CreateOne_Req): Promise<Result<OrgGroups_CreateOne_Res, Error>> {
        const { orgId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/orgs/${orgId}/groups`, payload)).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Rename a group, or change its description or accounts. Its members get the accounts from their next certificate.
     */
    async updateOne(request: OrgGroups_UpdateOne_Req): Promise<Result<OrgGroups_UpdateOne_Res, Error>> {
        const { orgId, groupId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/groups/${groupId}`, payload)).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a group. Its members lose its accounts from their next certificate.
     */
    async deleteOne(request: OrgGroups_DeleteOne_Req): Promise<Result<OrgGroups_DeleteOne_Res, Error>> {
        const { orgId, groupId } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/orgs/${orgId}/groups/${groupId}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Replace the members of a group. One not in the organization, or removed from it, is not found.
     */
    async setMembers(request: OrgGroups_SetMembers_Req): Promise<Result<OrgGroups_SetMembers_Res, Error>> {
        const { orgId, groupId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/groups/${groupId}/members`, payload)).pipe(
                map(this.validator.setMembers),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
