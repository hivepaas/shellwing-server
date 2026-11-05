import type { EOrgMemberStatus, OrgMember } from "~/organization/domain";

import type { EOrgRole } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * GET /orgs/{orgID}/members: sorted by username, not paged. Owners and admins get every member, removed ones included.
 */
export type OrgMembers_FindMany_Req = ApiRequestBase<{
    orgId: string;
}>;
export type OrgMembers_FindMany_Res = ApiResponseBase<OrgMember[]>;

export type OrgMembers_UpdateRole_Payload = {
    role: EOrgRole;
};
export type OrgMembers_UpdateRole_Req = ApiRequestBase<{
    orgId: string;
    userId: string;
    payload: OrgMembers_UpdateRole_Payload;
}>;
export type OrgMembers_UpdateRole_Res = ApiResponseBase<{ type: "success" }>;

export type OrgMembers_UpdateStatus_Payload = {
    status: EOrgMemberStatus;
};
export type OrgMembers_UpdateStatus_Req = ApiRequestBase<{
    orgId: string;
    userId: string;
    payload: OrgMembers_UpdateStatus_Payload;
}>;
export type OrgMembers_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;
