import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SshPrincipals_FindOne_Req = ApiRequestBase<{
    orgId: string;
    userId: string;
}>;
export type SshPrincipals_FindOne_Res = ApiResponseBase<string[]>;

export type SshPrincipals_UpdateOne_Payload = {
    /** Every account the member may log in as, in place of what they had. Empty takes them all away. */
    principals: string[];
};
export type SshPrincipals_UpdateOne_Req = ApiRequestBase<{
    orgId: string;
    userId: string;
    payload: SshPrincipals_UpdateOne_Payload;
}>;
/** What was set: sorted, without repeats. */
export type SshPrincipals_UpdateOne_Res = ApiResponseBase<string[]>;
