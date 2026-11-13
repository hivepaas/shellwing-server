import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { SshKrlToken_CreateOne_Res, SshKrlToken_DeleteOne_Res } from "./ssh-krl.api.contracts";

const CreateOneSchema = z.object({
    data: z.object({
        token: z.string(),
        createdAt: z.coerce.date(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class SshKrlApiValidator {
    createOne = (response: ApiHttpResponse): SshKrlToken_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    deleteOne = (response: ApiHttpResponse): SshKrlToken_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
