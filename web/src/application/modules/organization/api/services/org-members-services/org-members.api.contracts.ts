import type { OrgMember } from "~/organization/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * GET /orgs/{orgID}/members: for the organization's owners and admins. Sorted by username, not paged.
 */
export type OrgMembers_FindMany_Req = ApiRequestBase<{
    orgId: string;
}>;
export type OrgMembers_FindMany_Res = ApiResponseBase<OrgMember[]>;
