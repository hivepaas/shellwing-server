import type { EOrgRole } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export const EOrgMemberStatus = {
    Active: "active",
    Invited: "invited",
    Disabled: "disabled",
} as const;

export type EOrgMemberStatus = (typeof EOrgMemberStatus)[keyof typeof EOrgMemberStatus];

/**
 * A member of the organization, as its admins see them.
 */
export interface OrgMember {
    userId: string;
    username: string;
    email: string;
    fullName: string;
    role: OpenApiConstant<EOrgRole>;
    status: OpenApiConstant<EOrgMemberStatus>;

    /** The accounts their SSH certificates may log in as. Empty: none. */
    sshPrincipals: string[];
}
