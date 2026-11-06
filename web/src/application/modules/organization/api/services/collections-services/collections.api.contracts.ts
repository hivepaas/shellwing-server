import type { Collection, CollectionDetail, CollectionGrant, ECollectionPermission } from "~/organization/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * GET /orgs/{orgID}/collections: not paged. An owner or admin sees every collection of the organization.
 */
export type Collections_FindMany_Req = ApiRequestBase<{
    orgId: string;
}>;
export type Collections_FindMany_Res = ApiResponseBase<Collection[]>;

/**
 * GET /orgs/{orgID}/collections/{collectionID}: the collection, with its grants and who has access for those who manage
 * its access.
 */
export type Collections_FindOne_Req = ApiRequestBase<{
    orgId: string;
    collectionId: string;
}>;
export type Collections_FindOne_Res = ApiResponseBase<CollectionDetail>;

export type Collections_Payload = {
    /** 1 to 100 characters, trimmed; unique in the organization whatever its case (ERR_ALREADY_EXIST). */
    name: string;
    /** Up to 500 characters. */
    description: string;
};

export type Collections_CreateOne_Req = ApiRequestBase<{
    orgId: string;
    payload: Collections_Payload;
}>;
/** The collection made, with no key yet. An admin who is not an owner is granted Manage on it. */
export type Collections_CreateOne_Res = ApiResponseBase<Collection>;

export type Collections_UpdateOne_Req = ApiRequestBase<{
    orgId: string;
    collectionId: string;
    payload: Collections_Payload;
}>;
export type Collections_UpdateOne_Res = ApiResponseBase<Collection>;

export type Collections_DeleteOne_Req = ApiRequestBase<{
    orgId: string;
    collectionId: string;
}>;
export type Collections_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

/** A grant to a person or to a group: exactly one of userId and groupId. */
export type Collections_GrantPayload =
    | { userId: string; groupId?: never; permission: ECollectionPermission }
    | { userId?: never; groupId: string; permission: ECollectionPermission };

export type Collections_SetGrants_Payload = {
    /**
     * Every grant of the collection, in place of what it had. A person must be a member of the organization, not removed;
     * a group one of its groups.
     */
    grants: Collections_GrantPayload[];
};
export type Collections_SetGrants_Req = ApiRequestBase<{
    orgId: string;
    collectionId: string;
    payload: Collections_SetGrants_Payload;
}>;
/** The grants as they are now. */
export type Collections_SetGrants_Res = ApiResponseBase<CollectionGrant[]>;
