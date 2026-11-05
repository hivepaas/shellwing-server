import type { ColumnDef } from "@tanstack/react-table";
import type { OrgMember } from "~/organization/domain";
import { OrgGroupBadges, OrgMemberStatusBadge } from "~/organization/module-shared/components";
import type { OrgMemberAbilities } from "~/organization/module-shared/utils";

import { MemberMenuCell, MemberPersonCell, MemberRoleCell } from "./building-blocks";

interface ColumnsParams {
    orgId: string;
    actorIsOwner: boolean;
    getAbilities: (member: OrgMember) => OrgMemberAbilities;
}

function createColumns({ orgId, actorIsOwner, getAbilities }: ColumnsParams): ColumnDef<OrgMember>[] {
    return [
        {
            id: "person",
            accessorFn: member => member.fullName || member.username,
            header: "Person",
            enableSorting: true,
            cell: ({ row: { original } }) => (
                <MemberPersonCell
                    member={original}
                    isSelf={getAbilities(original).isSelf}
                />
            ),
        },
        {
            accessorKey: "email",
            header: "Email",
            enableSorting: true,
            cell: ({ row: { original } }) =>
                original.email ? <span>{original.email}</span> : <span className="text-muted-foreground">-</span>,
        },
        {
            accessorKey: "role",
            header: "Role",
            enableSorting: true,
            cell: ({ row: { original } }) => (
                <MemberRoleCell
                    orgId={orgId}
                    member={original}
                    abilities={getAbilities(original)}
                />
            ),
        },
        {
            accessorKey: "status",
            header: "Status",
            enableSorting: true,
            cell: ({ row: { original } }) => (
                <OrgMemberStatusBadge
                    status={original.status}
                    accountStatus={original.accountStatus}
                />
            ),
        },
        {
            id: "groups",
            header: "Groups",
            enableSorting: false,
            cell: ({ row: { original } }) => <OrgGroupBadges groups={original.groups} />,
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <MemberMenuCell
                    orgId={orgId}
                    member={original}
                    abilities={getAbilities(original)}
                    actorIsOwner={actorIsOwner}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const MembersTableDefs = Object.freeze({
    columns: createColumns,
});
