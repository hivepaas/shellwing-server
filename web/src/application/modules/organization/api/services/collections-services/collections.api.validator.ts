import { z } from "zod";
import { ECollectionPermission } from "~/organization/domain";

import { EOrgRole } from "@application/shared/enums";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    Collections_CreateOne_Res,
    Collections_DeleteOne_Res,
    Collections_FindMany_Res,
    Collections_FindOne_Res,
    Collections_SetGrants_Res,
    Collections_UpdateOne_Res,
} from "./collections.api.contracts";

const optionalString = z
    .string()
    .nullish()
    .transform(value => value ?? "");

const PermissionSchema = z.union([z.nativeEnum(ECollectionPermission), z.string()]);

const CollectionSchema = z.object({
    id: z.string(),
    name: z.string(),
    description: optionalString,
    keyVersion: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    // Empty for an admin with no grant.
    permission: PermissionSchema.nullish().transform(value => value ?? ""),
    hasKey: z
        .boolean()
        .nullish()
        .transform(value => value ?? false),
    // For those who manage the collection's access only.
    waiting: z
        .number()
        .nullish()
        .transform(value => value ?? 0),
    needsNewKey: z
        .boolean()
        .nullish()
        .transform(value => value ?? false),
});

const CollectionGrantSchema = z.object({
    userId: optionalString,
    username: optionalString,
    fullName: optionalString,
    groupId: optionalString,
    groupName: optionalString,
    permission: PermissionSchema,
});

const CollectionMemberIdentitySchema = z.object({
    x25519Public: z.string(),
    ed25519Public: z.string(),
    signature: z.string(),
    fingerprint: z.string(),
    updatedAt: z.coerce.date(),
});

const CollectionMemberSchema = z.object({
    userId: z.string(),
    username: z.string(),
    fullName: optionalString,
    role: z.union([z.nativeEnum(EOrgRole), z.string()]),
    permission: PermissionSchema,
    served: z
        .boolean()
        .nullish()
        .transform(value => value ?? true),
    hasKey: z
        .boolean()
        .nullish()
        .transform(value => value ?? false),
    identity: CollectionMemberIdentitySchema.nullish().transform(value => value ?? null),
});

const FindManySchema = z.object({
    data: z
        .array(CollectionSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

const FindOneSchema = z.object({
    data: z.object({
        collection: CollectionSchema,
        // Empty for those who do not manage the collection's access.
        grants: z
            .array(CollectionGrantSchema)
            .nullish()
            .transform(value => value ?? []),
        members: z
            .array(CollectionMemberSchema)
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const OneSchema = z.object({
    data: CollectionSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const GrantsSchema = z.object({
    data: z
        .array(CollectionGrantSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class CollectionsApiValidator {
    findMany = (response: ApiHttpResponse): Collections_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return { data, meta };
    };

    findOne = (response: ApiHttpResponse): Collections_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneSchema,
        });

        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): Collections_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: OneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): Collections_UpdateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: OneSchema,
        });

        return { data, meta };
    };

    deleteOne = (response: ApiHttpResponse): Collections_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    setGrants = (response: ApiHttpResponse): Collections_SetGrants_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: GrantsSchema,
        });

        return { data, meta };
    };
}
