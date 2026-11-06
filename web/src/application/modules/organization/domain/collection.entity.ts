import type { EOrgRole } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

export const ECollectionPermission = {
    /** See the collection's hosts and connect to them. */
    Connect: "connect",
    /** Also add and change its hosts. */
    Edit: "edit",
    /** Also grant others access and replace its key. */
    Manage: "manage",
} as const;

export type ECollectionPermission = (typeof ECollectionPermission)[keyof typeof ECollectionPermission];

/**
 * A space the organization shares: its hosts are sealed with a key the server and the web never hold. The shellwing app
 * of somebody holding the key gives it to each person granted access.
 */
export interface Collection {
    id: string;
    /** Unique in the organization, whatever its case. */
    name: string;
    /** Empty when none. */
    description: string;

    /** The version of the key its hosts are sealed with now: 0 until the app of somebody with Manage makes the first. */
    keyVersion: number;

    /** The viewer's own: empty for an admin with no grant. */
    permission: OpenApiConstant<ECollectionPermission>;
    /** Whether the viewer holds its current key. */
    hasKey: boolean;

    /** How many people with access are not given the current key yet. 0 for those who do not manage its access. */
    waiting: number;
    /** Somebody holding the current key has lost access: the app of somebody with Manage replaces it. */
    needsNewKey: boolean;
}

/**
 * A permission on a collection, granted to a person or to a group: one of userId and groupId.
 */
export interface CollectionGrant {
    /** Empty for a group's grant. */
    userId: string;
    username: string;
    fullName: string;

    /** Empty for a person's grant. */
    groupId: string;
    groupName: string;

    permission: OpenApiConstant<ECollectionPermission>;
}

/**
 * The public halves of the identity a person's shellwing app made for them: what the key is given to.
 */
export interface CollectionMemberIdentity {
    /** Base64 */
    x25519Public: string;
    /** Base64 */
    ed25519Public: string;
    /** Base64 */
    signature: string;
    /** Six groups of five digits, compared before the key is given. */
    fingerprint: string;
    updatedAt: Date;
}

/**
 * A person with access to a collection, through their role (an owner has Manage on every collection), their grant or
 * their groups'.
 */
export interface CollectionMember {
    userId: string;
    username: string;
    /** Empty when unknown. */
    fullName: string;
    role: OpenApiConstant<EOrgRole>;
    permission: OpenApiConstant<ECollectionPermission>;

    /** False for a suspended member or a disabled account: they keep their access, and are given nothing. */
    served: boolean;
    /** Whether they hold the collection's current key. */
    hasKey: boolean;

    /** None until their shellwing app publishes one: nobody can give them the key before. */
    identity: CollectionMemberIdentity | null;
}

/**
 * A collection, with its grants and who has access. The grants and the people are empty for those who do not manage
 * its access.
 */
export interface CollectionDetail {
    collection: Collection;
    grants: CollectionGrant[];
    members: CollectionMember[];
}
