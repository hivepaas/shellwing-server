import type { ColumnDef } from "@tanstack/react-table";
import type { OrgMember } from "~/organization/domain";
import { OrgMemberStatusBadge, OrgRoleBadge } from "~/organization/module-shared/components";
import type { GroupSshPrincipal } from "~/organization/module-shared/utils";

import { SshPrincipalsCell, SshPrincipalsEditCell } from "./building-blocks";

interface ColumnsParams {
    orgId: string;
    /** The accounts a member gets from their groups, besides their own. */
    getGroupPrincipals: (member: OrgMember) => GroupSshPrincipal[];
}

function createColumns({ orgId, getGroupPrincipals }: ColumnsParams): ColumnDef<OrgMember>[] {
    return [
        {
            accessorKey: "username",
            header: "Username",
            enableSorting: true,
            cell: ({ row: { original } }) => <span className="font-medium">{original.username}</span>,
        },
        {
            id: "person",
            accessorFn: member => member.fullName || member.email,
            header: "Name",
            enableSorting: true,
            cell: ({ row: { original } }) => (
                <div className="flex flex-col">
                    {original.fullName ? <span>{original.fullName}</span> : null}
                    {original.email ? <span className="text-xs text-muted-foreground">{original.email}</span> : null}
                    {!original.fullName && !original.email ? "-" : null}
                </div>
            ),
        },
        {
            accessorKey: "role",
            header: "Role",
            enableSorting: true,
            cell: ({ row: { original } }) => <OrgRoleBadge role={original.role} />,
        },
        {
            accessorKey: "status",
            header: "Status",
            enableSorting: true,
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <OrgMemberStatusBadge
                    status={original.status}
                    accountStatus={original.accountStatus}
                    className="justify-center"
                />
            ),
        },
        {
            id: "sshPrincipals",
            header: "May log in as",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <SshPrincipalsCell
                    principals={original.sshPrincipals}
                    groupPrincipals={getGroupPrincipals(original)}
                />
            ),
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <SshPrincipalsEditCell
                    orgId={orgId}
                    member={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const SshAccessTableDefs = Object.freeze({
    columns: createColumns,
});
