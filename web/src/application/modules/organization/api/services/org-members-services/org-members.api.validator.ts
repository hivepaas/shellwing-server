import { z } from "zod";
import { EOrgMemberStatus } from "~/organization/domain";

import { EOrgRole } from "@application/shared/enums";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { OrgMembers_FindMany_Res } from "./org-members.api.contracts";

const OrgMemberSchema = z.object({
    userId: z.string(),
    username: z.string(),
    email: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    fullName: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    role: z.union([z.nativeEnum(EOrgRole), z.string()]),
    status: z.union([z.nativeEnum(EOrgMemberStatus), z.string()]),
    sshPrincipals: z
        .array(z.string())
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

export class OrgMembersApiValidator {
    findMany = (response: ApiHttpResponse): OrgMembers_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return { data, meta };
    };
}
