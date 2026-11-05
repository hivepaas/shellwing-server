import { EOrgMemberStatus, type OrgMember } from "~/organization/domain";

import { EOrgRole, EUserStatus } from "@application/shared/enums";

/**
 * Who is acting on the members: the signed-in user, by their id and their role in the organization.
 */
export interface OrgMemberActor {
    userId: string;
    role: string;
}

/**
 * What the actor may do to a member, by the server's rules. The server checks them again and has the last word.
 */
export interface OrgMemberAbilities {
    isSelf: boolean;
    /** The organization's only active owner with an active account: they cannot be demoted, suspended or removed. */
    isLastActiveOwner: boolean;
    canChangeRole: boolean;
    /** Owner is offered only to an owner. */
    canMakeOwner: boolean;
    canSuspend: boolean;
    canReactivate: boolean;
    canRemove: boolean;
    canAddBack: boolean;
    canLeave: boolean;
}

function isActiveOwner(member: OrgMember): boolean {
    return (
        member.role === EOrgRole.Owner &&
        member.status === EOrgMemberStatus.Active &&
        member.accountStatus === EUserStatus.Active
    );
}

/**
 * Owners and admins change members and admins; only an owner changes an owner, or makes one. Anybody may leave. The last
 * active owner stays one: somebody else must be made an owner first.
 */
export function getOrgMemberAbilities(
    actor: OrgMemberActor,
    member: OrgMember,
    members: readonly OrgMember[],
): OrgMemberAbilities {
    const isSelf = member.userId === actor.userId;
    const isActorOwner = actor.role === EOrgRole.Owner;
    const isActorAdmin = isActorOwner || actor.role === EOrgRole.Admin;
    const isRemoved = member.status === EOrgMemberStatus.Removed;

    // An owner stays one, and an active owner stays active, unless somebody else is an active owner.
    const isOnlyOwner =
        member.role === EOrgRole.Owner &&
        !members.some(other => other.userId !== member.userId && isActiveOwner(other));
    const isLastActiveOwner = isOnlyOwner && member.status === EOrgMemberStatus.Active;

    // Who may be acted on at all: an owner only by an owner.
    const mayManage = isActorAdmin && (member.role !== EOrgRole.Owner || isActorOwner);

    return {
        isSelf,
        isLastActiveOwner,
        canChangeRole: mayManage && !isRemoved && !isOnlyOwner,
        canMakeOwner: isActorOwner,
        canSuspend: mayManage && !isSelf && member.status === EOrgMemberStatus.Active && !isLastActiveOwner,
        canReactivate: mayManage && !isSelf && member.status === EOrgMemberStatus.Suspended,
        canRemove: mayManage && !isSelf && !isRemoved && !isLastActiveOwner,
        canAddBack: mayManage && !isSelf && isRemoved,
        canLeave: isSelf && !isRemoved && !isLastActiveOwner,
    };
}
