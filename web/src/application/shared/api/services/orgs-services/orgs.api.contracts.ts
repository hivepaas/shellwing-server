import type { Org } from "@application/shared/entities";

import type { ApiResponseBase } from "@infrastructure/api";

/**
 * List the organizations the signed-in user is a member of
 */
export type Orgs_FindMany_Res = ApiResponseBase<Org[]>;
