import { z } from "zod";
import { type UserIdentity_FindMine_Res } from "~/user-management/api/services/users-services/user-identity/user-identity.api.contracts";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * User identity schema
 */
const UserIdentitySchema = z.object({
    x25519Public: z.string(),
    ed25519Public: z.string(),
    signature: z.string(),
    fingerprint: z.string(),
    updatedAt: z.coerce.date(),
});

/**
 * Find my identity API response schema
 */
const FindMineSchema = z.object({
    data: UserIdentitySchema.nullable(),
    meta: BaseMetaApiSchema.nullish(),
});

export class UserIdentityApiValidator {
    /**
     * Validate and transform find my identity API response
     */
    findMine = (response: ApiHttpResponse): UserIdentity_FindMine_Res => {
        const { data } = parseApiResponse({
            response,
            schema: FindMineSchema,
        });

        return {
            data: data
                ? {
                      x25519Public: data.x25519Public,
                      ed25519Public: data.ed25519Public,
                      signature: data.signature,
                      fingerprint: data.fingerprint,
                      updatedAt: data.updatedAt,
                  }
                : null,
        };
    };
}
