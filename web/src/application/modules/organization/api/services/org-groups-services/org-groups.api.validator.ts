import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    OrgGroups_CreateOne_Res,
    OrgGroups_DeleteOne_Res,
    OrgGroups_FindMany_Res,
    OrgGroups_SetMembers_Res,
    OrgGroups_UpdateOne_Res,
} from "./org-groups.api.contracts";

const OrgGroupMemberSchema = z.object({
    userId: z.string(),
    username: z.string(),
    fullName: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
});

const OrgGroupSchema = z.object({
    id: z.string(),
    name: z.string(),
    description: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    sshPrincipals: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    members: z
        .array(OrgGroupMemberSchema)
        .nullish()
        .transform(value => value ?? []),
});

const FindManySchema = z.object({
    data: z
        .array(OrgGroupSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

const OneSchema = z.object({
    data: OrgGroupSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class OrgGroupsApiValidator {
    findMany = (response: ApiHttpResponse): OrgGroups_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): OrgGroups_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: OneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): OrgGroups_UpdateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: OneSchema,
        });

        return { data, meta };
    };

    deleteOne = (response: ApiHttpResponse): OrgGroups_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    setMembers = (response: ApiHttpResponse): OrgGroups_SetMembers_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: OneSchema,
        });

        return { data, meta };
    };
}
