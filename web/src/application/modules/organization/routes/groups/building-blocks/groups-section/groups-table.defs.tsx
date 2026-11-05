import type { ColumnDef } from "@tanstack/react-table";
import type { OrgGroup } from "~/organization/domain";
import { SshPrincipalBadges } from "~/organization/module-shared/components";

import { GroupMembersCell, GroupMenuCell, GroupNameCell } from "./building-blocks";

function createColumns(orgId: string): ColumnDef<OrgGroup>[] {
    return [
        {
            accessorKey: "name",
            header: "Group",
            enableSorting: true,
            cell: ({ row: { original } }) => <GroupNameCell group={original} />,
        },
        {
            id: "sshPrincipals",
            header: "SSH accounts",
            enableSorting: false,
            cell: ({ row: { original } }) => <SshPrincipalBadges principals={original.sshPrincipals} />,
        },
        {
            id: "members",
            accessorFn: group => group.members.length,
            header: "Members",
            enableSorting: true,
            cell: ({ row: { original } }) => <GroupMembersCell members={original.members} />,
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <GroupMenuCell
                    orgId={orgId}
                    group={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const GroupsTableDefs = Object.freeze({
    columns: createColumns,
});
