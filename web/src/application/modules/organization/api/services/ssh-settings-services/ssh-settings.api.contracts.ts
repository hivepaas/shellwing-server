import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SshSettings_UpdateOne_Payload = {
    /** How long a certificate lasts: 5 to 1440 minutes. */
    certTtlMinutes: number;
};
export type SshSettings_UpdateOne_Req = ApiRequestBase<{
    orgId: string;
    payload: SshSettings_UpdateOne_Payload;
}>;
export type SshSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
