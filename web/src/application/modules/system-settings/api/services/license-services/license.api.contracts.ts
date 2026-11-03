import type { License } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type License_FindOne_Res = ApiResponseBase<License>;

export type License_InstallOne_Payload = {
    /** The key as it was pasted: wrapped or indented is the same key. */
    key: string;
};
export type License_InstallOne_Req = ApiRequestBase<{
    payload: License_InstallOne_Payload;
}>;
export type License_InstallOne_Res = ApiResponseBase<License>;

export type License_RemoveOne_Res = ApiResponseBase<License>;
