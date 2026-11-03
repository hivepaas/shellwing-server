import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { SshSettings_UpdateOne_Res } from "./ssh-settings.api.contracts";

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class SshSettingsApiValidator {
    updateOne = (response: ApiHttpResponse): SshSettings_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
