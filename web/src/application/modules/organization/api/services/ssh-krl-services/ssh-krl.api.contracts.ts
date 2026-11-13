import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/** A token a target server fetches the revocation list with. It is in this answer and nowhere else. */
export interface SshKrlToken {
    token: string;
    createdAt: Date;
}

export type SshKrlToken_CreateOne_Req = ApiRequestBase<{
    orgId: string;
}>;
export type SshKrlToken_CreateOne_Res = ApiResponseBase<SshKrlToken>;

export type SshKrlToken_DeleteOne_Req = ApiRequestBase<{
    orgId: string;
}>;
export type SshKrlToken_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
