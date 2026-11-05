import type { OrgGroup, OrgMember } from "~/organization/domain";

/**
 * An account a member gets from their groups, with the groups that grant it.
 */
export interface GroupSshPrincipal {
    name: string;
    groupNames: string[];
}

/**
 * The accounts a member gets from their groups and has not of their own: what their certificates may log in as besides
 * their own accounts.
 */
export function getGroupSshPrincipals(
    member: OrgMember,
    groupsById: ReadonlyMap<string, OrgGroup>,
): GroupSshPrincipal[] {
    const own = new Set(member.sshPrincipals);
    const byName = new Map<string, GroupSshPrincipal>();

    for (const ref of member.groups) {
        const group = groupsById.get(ref.id);

        if (group === undefined) {
            continue;
        }

        for (const name of group.sshPrincipals) {
            if (own.has(name)) {
                continue;
            }

            const found = byName.get(name);

            if (found === undefined) {
                byName.set(name, { name, groupNames: [group.name] });
            } else {
                found.groupNames.push(group.name);
            }
        }
    }

    return [...byName.values()];
}
