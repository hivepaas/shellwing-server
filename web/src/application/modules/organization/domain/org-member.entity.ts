import type { EOrgRole, EUserStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export const EOrgMemberStatus = {
    Active: "active",
    /** The organization gives them nothing for now, and keeps their role and accounts. */
    Suspended: "suspended",
    /** Gone from the organization: their role is member again and their accounts are cleared. */
    Removed: "removed",
} as const;

export type EOrgMemberStatus = (typeof EOrgMemberStatus)[keyof typeof EOrgMemberStatus];

/**
 * A member of the organization.
 *
 * The organization's owners and admins see everything of every member, removed ones included. A plain member sees the
 * active members by name and role only: for them the rest stands at its default - an empty email and accounts, an
 * active status and account - which is true of whom they see.
 */
export interface OrgMember {
    userId: string;
    username: string;
    fullName: string;
    role: OpenApiConstant<EOrgRole>;

    /** Empty when unknown. */
    email: string;

    /** Their status in the organization: the organization's say. */
    status: OpenApiConstant<EOrgMemberStatus>;

    /** Their user account's own status, on this server: pending until they accept their invite. */
    accountStatus: OpenApiConstant<EUserStatus>;

    /** The accounts their SSH certificates may log in as. Empty: none. */
    sshPrincipals: string[];
}
