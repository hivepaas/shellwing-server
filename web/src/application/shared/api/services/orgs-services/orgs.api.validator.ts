import { z } from "zod";

import { EOrgRole } from "@application/shared/enums";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { Orgs_FindMany_Res } from "./orgs.api.contracts";

const OrgSchema = z.object({
    id: z.string(),
    slug: z.string(),
    name: z.string(),
    role: z.union([z.nativeEnum(EOrgRole), z.string()]),
    sshCertTtlMinutes: z.number(),
    sshPrincipals: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    sshCertExtensions: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    // Left out while its servers have no token of their own.
    sshKrlTokenCreatedAt: z.coerce
        .date()
        .nullish()
        .transform(value => value ?? null),
});

const FindManySchema = z.object({
    data: z
        .array(OrgSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

export class OrgsApiValidator {
    findMany = (response: ApiHttpResponse): Orgs_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return { data, meta };
    };
}
