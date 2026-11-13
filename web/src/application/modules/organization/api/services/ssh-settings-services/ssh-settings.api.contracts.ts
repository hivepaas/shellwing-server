import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * Each field is optional, and left out is left alone: a form changes its own setting without writing back the other
 * one as it read it some time ago. An empty list of extensions is a setting of its own - a certificate that logs in
 * and runs a command, with no terminal.
 */
export type SshSettings_UpdateOne_Payload = {
    /** How long a certificate lasts: 5 to 1440 minutes. */
    certTtlMinutes?: number;
    /** What a certificate is allowed to do, by sshd's extension names. */
    certExtensions?: string[];
};
export type SshSettings_UpdateOne_Req = ApiRequestBase<{
    orgId: string;
    payload: SshSettings_UpdateOne_Payload;
}>;
export type SshSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
