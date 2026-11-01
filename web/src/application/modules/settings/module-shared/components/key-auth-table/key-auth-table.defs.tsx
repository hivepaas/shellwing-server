import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingKeyAuth } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { KeyAuthEditCell, KeyAuthMenuCell } from "./building-blocks";

function createColumns(): ColumnDef<SettingKeyAuth>[] {
    return [
        {
            id: "view",
            accessorKey: "inherited",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => <KeyAuthEditCell id={original.id} />,
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            accessorKey: "keyId",
            header: "Key ID",
            enableSorting: true,
            cell: ({ row: { original } }) =>
                original.keyId ? (
                    <Badge
                        tone="neutral"
                        className="font-mono"
                    >
                        {original.keyId}
                    </Badge>
                ) : (
                    "-"
                ),
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <SettingStatusBadge status={original.status} />
                </div>
            ),
        },
        {
            accessorKey: "expireAt",
            header: "Expire At",
            cell: ({ row: { original } }) => {
                if (!original.expireAt) {
                    return "-";
                }

                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => <KeyAuthMenuCell keyAuth={original} />,
            meta: {
                align: "right",
            },
        },
    ];
}

export const KeyAuthTableDefs = Object.freeze({
    columns: createColumns,
});
