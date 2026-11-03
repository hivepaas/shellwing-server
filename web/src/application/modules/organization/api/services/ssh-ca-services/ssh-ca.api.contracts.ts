import type { SshCa } from "~/organization/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SshCas_FindMany_Req = ApiRequestBase<{
    orgId: string;
}>;
export type SshCas_FindMany_Res = ApiResponseBase<SshCa[]>;

export type SshCas_CreateOne_Req = ApiRequestBase<{
    orgId: string;
}>;
export type SshCas_CreateOne_Res = ApiResponseBase<SshCa>;

export type SshCas_DeleteOne_Req = ApiRequestBase<{
    orgId: string;
    caId: string;
}>;
export type SshCas_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
