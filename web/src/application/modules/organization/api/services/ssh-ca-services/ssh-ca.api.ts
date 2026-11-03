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
     * Make a new certificate authority. One already active is retired, and its private key destroyed.
     */
    async createOne(request: SshCas_CreateOne_Req): Promise<Result<SshCas_CreateOne_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/orgs/${orgId}/ssh/cas`)).pipe(
                map(this.validator.createOne),
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
