import { z } from "zod";
import { ELicenseStatus } from "~/system-settings/domain";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { License_FindOne_Res, License_InstallOne_Res, License_RemoveOne_Res } from "./license.api.contracts";

/** The server's maxUsers of a key that sets no limit. */
const NO_USER_LIMIT = -1;

const LicenseSchema = z
    .object({
        status: z.union([z.nativeEnum(ELicenseStatus), z.string()]),
        installationId: z.string(),
        plan: z
            .string()
            .nullish()
            .transform(value => value ?? ""),
        maxUsers: z.number(),
        freeUsers: z.number(),
        usedUsers: z.number(),
        trial: z
            .boolean()
            .nullish()
            .transform(value => value ?? false),
        customerName: z
            .string()
            .nullish()
            .transform(value => value ?? ""),
        // Left out when there is none.
        expiresAt: z.coerce
            .date()
            .nullish()
            .transform(value => value ?? null),
        graceEndsAt: z.coerce
            .date()
            .nullish()
            .transform(value => value ?? null),
    })
    .transform(license => ({
        ...license,
        maxUsers: license.maxUsers === NO_USER_LIMIT ? null : license.maxUsers,
    }));

const LicenseResponseSchema = z.object({
    data: LicenseSchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class LicenseApiValidator {
    findOne = (response: ApiHttpResponse): License_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: LicenseResponseSchema,
        });

        return { data, meta };
    };

    installOne = (response: ApiHttpResponse): License_InstallOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: LicenseResponseSchema,
        });

        return { data, meta };
    };

    removeOne = (response: ApiHttpResponse): License_RemoveOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: LicenseResponseSchema,
        });

        return { data, meta };
    };
}
