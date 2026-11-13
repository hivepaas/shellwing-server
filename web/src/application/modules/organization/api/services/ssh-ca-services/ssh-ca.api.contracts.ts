import type { ESshCaStatus, SshCa } from "~/organization/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SshCas_FindMany_Req = ApiRequestBase<{
    orgId: string;
}>;
export type SshCas_FindMany_Res = ApiResponseBase<SshCa[]>;

export type SshCas_CreateOne_Req = ApiRequestBase<{
    orgId: string;
    /**
     * Retire the authority that signs, at once, and refuse every certificate it signed.
     *
     * Left out, the new authority is published instead: the servers are given its key, nothing is signed with it, and
     * the one signing carries on until it is promoted. That is rotating, and it locks nobody out.
     */
    replace?: boolean;
}>;
export type SshCas_CreateOne_Res = ApiResponseBase<SshCa>;

export type SshCas_SetStatus_Req = ApiRequestBase<{
    orgId: string;
    caId: string;
    /** "active" to sign with it from now on, or "retired" to stop trusting it. */
    status: ESshCaStatus;
}>;
export type SshCas_SetStatus_Res = ApiResponseBase<SshCa>;

export type SshCas_DeleteOne_Req = ApiRequestBase<{
    orgId: string;
    caId: string;
}>;
export type SshCas_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
