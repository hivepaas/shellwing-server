import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    Collections_CreateOne_Req,
    Collections_CreateOne_Res,
    Collections_DeleteOne_Req,
    Collections_DeleteOne_Res,
    Collections_FindMany_Req,
    Collections_FindMany_Res,
    Collections_FindOne_Req,
    Collections_FindOne_Res,
    Collections_SetGrants_Req,
    Collections_SetGrants_Res,
    Collections_UpdateOne_Req,
    Collections_UpdateOne_Res,
} from "./collections.api.contracts";
import type { CollectionsApiValidator } from "./collections.api.validator";

export class CollectionsApi extends BaseApi {
    public constructor(private readonly validator: CollectionsApiValidator) {
        super();
    }

    /**
     * The organization's collections, each with its key's state
     */
    async findMany(
        request: Collections_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<Collections_FindMany_Res, Error>> {
        const { orgId } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/orgs/${orgId}/collections`, {
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
     * A collection, with its grants and who has access
     */
    async findOne(
        request: Collections_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<Collections_FindOne_Res, Error>> {
        const { orgId, collectionId } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/orgs/${orgId}/collections/${collectionId}`, {
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
     * Make a collection, with no key. A name taken already, whatever its case, is refused (ERR_ALREADY_EXIST).
     */
    async createOne(request: Collections_CreateOne_Req): Promise<Result<Collections_CreateOne_Res, Error>> {
        const { orgId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/orgs/${orgId}/collections`, payload)).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Rename a collection, or change its description.
     */
    async updateOne(request: Collections_UpdateOne_Req): Promise<Result<Collections_UpdateOne_Res, Error>> {
        const { orgId, collectionId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/collections/${collectionId}`, payload)).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a collection. Nobody syncs it from then on.
     */
    async deleteOne(request: Collections_DeleteOne_Req): Promise<Result<Collections_DeleteOne_Res, Error>> {
        const { orgId, collectionId } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/orgs/${orgId}/collections/${collectionId}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Replace the grants of a collection. Those granted wait for the key until an app holding it gives it to them.
     */
    async setGrants(request: Collections_SetGrants_Req): Promise<Result<Collections_SetGrants_Res, Error>> {
        const { orgId, collectionId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/collections/${collectionId}/grants`, payload)).pipe(
                map(this.validator.setGrants),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
