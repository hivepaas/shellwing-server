import { z } from "zod";
import { ESshCaKeyProvider, ESshCaStatus } from "~/organization/domain";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SshCas_CreateOne_Res,
    SshCas_DeleteOne_Res,
    SshCas_FindMany_Res,
    SshCas_SetStatus_Res,
} from "./ssh-ca.api.contracts";

const SshCaSchema = z.object({
    id: z.string(),
    status: z.union([z.nativeEnum(ESshCaStatus), z.string()]),
    publicKey: z.string(),
    keyProvider: z
        .union([z.nativeEnum(ESshCaKeyProvider), z.string()])
        .nullish()
        .transform(value => value ?? ESshCaKeyProvider.Local),
    createdAt: z.coerce.date(),
    // Left out while the CA is active.
    retiredAt: z.coerce
        .date()
        .nullish()
        .transform(value => value ?? null),
});

const FindManySchema = z.object({
    data: z
        .array(SshCaSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: SshCaSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class SshCaApiValidator {
    findMany = (response: ApiHttpResponse): SshCas_FindMany_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManySchema,
        });

        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): SshCas_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    setStatus = (response: ApiHttpResponse): SshCas_SetStatus_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    deleteOne = (response: ApiHttpResponse): SshCas_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
