import { z } from "zod";
import { EOrgMemberStatus } from "~/organization/domain";

import { EOrgRole, EUserStatus } from "@application/shared/enums";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    OrgMembers_FindMany_Res,
    OrgMembers_UpdateRole_Res,
    OrgMembers_UpdateStatus_Res,
} from "./org-members.api.contracts";

const OrgMemberSchema = z.object({
    userId: z.string(),
    username: z.string(),
    fullName: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    role: z.union([z.nativeEnum(EOrgRole), z.string()]),
    // For the organization's owners and admins only. A plain member sees only active members with active accounts.
    email: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    status: z
        .union([z.nativeEnum(EOrgMemberStatus), z.string()])
        .nullish()
        .transform(value => value ?? EOrgMemberStatus.Active),
    accountStatus: z
        .union([z.nativeEnum(EUserStatus), z.string()])
        .nullish()
        .transform(value => value ?? EUserStatus.Active),
    sshPrincipals: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    // For the organization's owners and admins only.
    groups: z
        .array(
            z.object({
                id: z.string(),
                name: z.string(),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
});

const FindManySchema = z.object({
    data: z
        .array(OrgMemberSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class OrgMembersApiValidator {
    findMany = (response: ApiHttpResponse): OrgMembers_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return { data, meta };
    };

    updateRole = (response: ApiHttpResponse): OrgMembers_UpdateRole_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): OrgMembers_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
