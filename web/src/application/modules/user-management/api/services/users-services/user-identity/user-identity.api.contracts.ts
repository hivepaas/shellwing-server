import { type UserIdentity } from "~/user-management/domain";

import { type ApiResponseBase } from "@infrastructure/api";

/**
 * Find the signed-in person's identity - null when their app has published none yet
 */
export type UserIdentity_FindMine_Res = ApiResponseBase<UserIdentity | null>;
