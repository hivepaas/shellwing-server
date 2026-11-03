import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { SshPrincipals_FindOne_Res, SshPrincipals_UpdateOne_Res } from "./ssh-principals.api.contracts";

const PrincipalsSchema = z.object({
    data: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

export class SshPrincipalsApiValidator {
    findOne = (response: ApiHttpResponse): SshPrincipals_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: PrincipalsSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): SshPrincipals_UpdateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: PrincipalsSchema,
        });

        return { data, meta };
    };
}
