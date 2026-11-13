import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SshCas_CreateOne_Req,
    SshCas_CreateOne_Res,
    SshCas_DeleteOne_Req,
    SshCas_DeleteOne_Res,
    SshCas_FindMany_Req,
    SshCas_FindMany_Res,
    SshCas_SetStatus_Req,
    SshCas_SetStatus_Res,
} from "./ssh-ca.api.contracts";
import type { SshCaApiValidator } from "./ssh-ca.api.validator";

export class SshCaApi extends BaseApi {
    public constructor(private readonly validator: SshCaApiValidator) {
        super();
    }

    /**
     * The organization's certificate authorities: the active one, and the retired ones not yet deleted
     */
    async findMany(request: SshCas_FindMany_Req, signal?: AbortSignal): Promise<Result<SshCas_FindMany_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/orgs/${orgId}/ssh/cas`, {
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
     * Make a new certificate authority: published beside the one that signs, or replacing it outright
     */
    async createOne(request: SshCas_CreateOne_Req): Promise<Result<SshCas_CreateOne_Res, Error>> {
        const { orgId, replace } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/orgs/${orgId}/ssh/cas`, { replace: replace ?? false })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Promote a published certificate authority so it signs, or retire it. Only a published one moves
     * (ERR_SSH_CA_NOT_PUBLISHED).
     */
    async setStatus(request: SshCas_SetStatus_Req): Promise<Result<SshCas_SetStatus_Res, Error>> {
        const { orgId, caId, status } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/ssh/cas/${caId}/status`, { status })).pipe(
                map(this.validator.setStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a retired certificate authority. The active one is refused (ERR_SSH_CA_ACTIVE).
     */
    async deleteOne(request: SshCas_DeleteOne_Req): Promise<Result<SshCas_DeleteOne_Res, Error>> {
        const { orgId, caId } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/orgs/${orgId}/ssh/cas/${caId}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
