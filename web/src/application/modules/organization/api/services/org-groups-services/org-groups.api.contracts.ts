import type { OrgGroup } from "~/organization/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * GET /orgs/{orgID}/groups: sorted by name, not paged. For the organization's owners and admins.
 */
export type OrgGroups_FindMany_Req = ApiRequestBase<{
    orgId: string;
}>;
export type OrgGroups_FindMany_Res = ApiResponseBase<OrgGroup[]>;

export type OrgGroups_Payload = {
    /** 1 to 100 characters, trimmed; unique in the organization whatever its case (ERR_ALREADY_EXIST). */
    name: string;
    /** Up to 500 characters. */
    description: string;
    /** Up to 32, by the rules of a member's own (ERR_SSH_PRINCIPAL_INVALID). Empty: none. */
    sshPrincipals: string[];
};

export type OrgGroups_CreateOne_Req = ApiRequestBase<{
    orgId: string;
    payload: OrgGroups_Payload;
}>;
/** The group made, with no members yet. */
export type OrgGroups_CreateOne_Res = ApiResponseBase<OrgGroup>;

export type OrgGroups_UpdateOne_Req = ApiRequestBase<{
    orgId: string;
    groupId: string;
    payload: OrgGroups_Payload;
}>;
export type OrgGroups_UpdateOne_Res = ApiResponseBase<OrgGroup>;

export type OrgGroups_DeleteOne_Req = ApiRequestBase<{
    orgId: string;
    groupId: string;
}>;
export type OrgGroups_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type OrgGroups_SetMembers_Payload = {
    /** Every member of the group, in place of who was. Each must be a member of the organization, not removed. */
    userIds: string[];
};
export type OrgGroups_SetMembers_Req = ApiRequestBase<{
    orgId: string;
    groupId: string;
    payload: OrgGroups_SetMembers_Payload;
}>;
export type OrgGroups_SetMembers_Res = ApiResponseBase<OrgGroup>;
