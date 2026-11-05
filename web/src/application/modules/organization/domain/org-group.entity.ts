/**
 * A member of a group, by name.
 */
export interface OrgGroupMember {
    userId: string;
    username: string;
    /** Empty when unknown. */
    fullName: string;
}

/**
 * A group, by its id and name: how a member's groups are listed with them.
 */
export interface OrgGroupRef {
    id: string;
    name: string;
}

/**
 * A group of the organization's members. Its SSH accounts are each member's, on top of their own.
 */
export interface OrgGroup {
    id: string;
    /** Unique in the organization, whatever its case. */
    name: string;
    /** Empty when none. */
    description: string;

    /** The accounts its members' certificates may log in as. Empty: none. */
    sshPrincipals: string[];

    /** By username. */
    members: OrgGroupMember[];
}
