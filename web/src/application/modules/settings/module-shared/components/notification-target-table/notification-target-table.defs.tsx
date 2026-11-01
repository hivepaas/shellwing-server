import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingNotification } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import {
    NotificationTargetBadgesCell,
    NotificationTargetEditCell,
    NotificationTargetMenuCell,
} from "./building-blocks";

function createColumns(): ColumnDef<SettingNotification>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => <NotificationTargetEditCell id={original.id} />,
            meta: { align: "center", titleAlign: "center" },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            id: "targets",
            header: "Targets",
            cell: ({ row: { original } }) => <NotificationTargetBadgesCell notification={original} />,
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: { align: "center", titleAlign: "center" },
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
                if (!original.expireAt) return "-";
                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => <NotificationTargetMenuCell notificationTarget={original} />,
            meta: { align: "right" },
        },
    ];
}

export const NotificationTargetTableDefs = Object.freeze({
    columns: createColumns,
});
